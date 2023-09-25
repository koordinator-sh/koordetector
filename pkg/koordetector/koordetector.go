/*
Copyright 2022 The Koordinator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package koordetector

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/koordinator-sh/koordinator/pkg/koordlet/util/system"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/koordinator-sh/koordetector/pkg/koordetector/config"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/metrics"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/metricsadvisor"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/statesinformer"
)

var (
	scheme = apiruntime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

type Daemon interface {
	Run(stopCh <-chan struct{})
}

type daemon struct {
	collector      metricsadvisor.MetricAdvisor
	statesInformer statesinformer.StatesInformer
}

func NewDaemon(config *config.Configuration) (Daemon, error) {
	// get node name
	nodeName := os.Getenv("NODE_NAME")
	if len(nodeName) == 0 {
		return nil, fmt.Errorf("failed to new daemon: NODE_NAME env is empty")
	}
	klog.Infof("NODE_NAME is %v,start time %v", nodeName, float64(time.Now().Unix()))
	metrics.RecordKoordetectorStartTime(nodeName, float64(time.Now().Unix()))

	kubeClient := clientset.NewForConfigOrDie(config.KubeRestConf)

	// only sync pod info
	statesInformer := statesinformer.NewStatesInformer(config.StatesInformerConf, kubeClient, nodeName)

	// setup cgroup path formatter from cgroup driver type
	var detectCgroupDriver system.CgroupDriverType
	if pollErr := wait.PollImmediate(time.Second*10, time.Minute, func() (bool, error) {
		driver := system.GuessCgroupDriverFromCgroupName()
		if driver.Validate() {
			detectCgroupDriver = driver
			return true, nil
		}
		klog.Infof("can not detect cgroup driver from 'kubepods' cgroup name")

		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, v1.GetOptions{})
		if err != nil || node == nil {
			klog.Error("Can't get node")
			return false, nil
		}

		port := int(node.Status.DaemonEndpoints.KubeletEndpoint.Port)
		if driver, err := system.GuessCgroupDriverFromKubeletPort(port); err == nil && driver.Validate() {
			detectCgroupDriver = driver
			return true, nil
		} else {
			klog.Errorf("guess kubelet cgroup driver failed, retry...: %v", err)
			return false, nil
		}
	}); pollErr != nil {
		return nil, fmt.Errorf("can not detect kubelet cgroup driver: %v", pollErr)
	}
	system.SetupCgroupPathFormatter(detectCgroupDriver)
	klog.Infof("Node %s use '%s' as cgroup driver", nodeName, string(detectCgroupDriver))

	// add metric collector
	collectorService := metricsadvisor.NewMetricAdvisor(config.CollectorConf, statesInformer)

	d := &daemon{
		collector:      collectorService,
		statesInformer: statesInformer,
	}
	return d, nil
}

func (d *daemon) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	klog.Infof("Starting daemon")

	// start states informer
	go func() {
		if err := d.statesInformer.Run(stopCh); err != nil {
			klog.Fatalf("Unable to run the states informer: ", err)
		}
	}()
	// wait for metric advisor sync
	if !cache.WaitForCacheSync(stopCh, d.statesInformer.HasSynced) {
		klog.Fatalf("time out waiting for states informer to sync")
	}

	// start metric advisor
	go func() {
		if err := d.collector.Run(stopCh); err != nil {
			klog.Fatalf("Unable to run the metric advisor: ", err)
		}
	}()

	// wait for metric advisor sync
	if !cache.WaitForCacheSync(stopCh, d.collector.HasSynced) {
		klog.Fatalf("time out waiting for metric advisor to sync")
	}

	klog.Info("Start daemon successfully")
	<-stopCh
	// close all gc goroutines for prometheus metrics
	metrics.StopAllExpireMetrics()
	klog.Info("Shutting down daemon")
}
