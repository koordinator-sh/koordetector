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
	"fmt"
	"os"
	"time"

	"github.com/koordinator-sh/koordinator/pkg/koordlet/metricsadvisor"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"

	"github.com/koordinator-sh/koordetector/pkg/koordetector/config"
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
	// metrics.RecordKoordletStartTime(nodeName, float64(time.Now().Unix()))

	kubeClient := clientset.NewForConfigOrDie(config.KubeRestConf)

	// only sync pod info
	statesInformerConf := statesinformer.NewDefaultConfig()

	statesInformer := statesinformer.NewStatesInformer(statesInformerConf, kubeClient, nodeName)

	// add metric collector

	d := &daemon{
		statesInformer: statesInformer,
	}
	return d, nil
}

func (d *daemon) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	klog.Infof("Starting daemon")

	klog.Info("Start daemon successfully")
	<-stopCh
	klog.Info("Shutting down daemon")
}
