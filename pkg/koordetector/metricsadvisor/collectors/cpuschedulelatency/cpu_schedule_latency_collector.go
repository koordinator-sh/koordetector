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

package cpuschedulelatency

import (
	"fmt"
	"os"
	"time"

	"github.com/koordinator-sh/koordinator/pkg/koordlet/util/system"
	"go.uber.org/atomic"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/koordinator-sh/koordetector/pkg/features"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/metrics"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/metricsadvisor/framework"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/statesinformer"
	csl "github.com/koordinator-sh/koordetector/pkg/koordetector/util/cpu_schedule_latency"
)

const (
	CollectorName = "CPUScheduleLatencyCollector"
)

type cslCollector struct {
	cslEnabled                bool
	cslCollectInterval        time.Duration
	collectTimeWindowDuration time.Duration

	started        *atomic.Bool
	statesInformer statesinformer.StatesInformer
}

func New(opt *framework.Options) framework.Collector {
	return &cslCollector{
		cslEnabled:                features.DefaultKoordetectorFeatureGate.Enabled(features.CSLCollector),
		cslCollectInterval:        time.Duration(opt.Config.CSLCollectorIntervalSeconds) * time.Second,
		collectTimeWindowDuration: time.Duration(opt.Config.CSLCollectorTimeWindowSeconds) * time.Second,

		started:        atomic.NewBool(false),
		statesInformer: opt.StatesInformer,
	}
}

func (c *cslCollector) Enabled() bool {
	return c.cslEnabled
}

func (c *cslCollector) Setup(s *framework.Context) {}

func (c *cslCollector) Started() bool {
	return c.started.Load()
}

func (c *cslCollector) Run(stopCh <-chan struct{}) {
	if !cache.WaitForCacheSync(stopCh, c.statesInformer.HasSynced) {
		// Koordetector exit because of statesInformer sync failed.
		klog.Fatalf("timed out waiting for states informer caches to sync")
	}
	if c.cslEnabled {
		go wait.Until(func() {
			c.collectContainerCSL()
		}, c.cslCollectInterval, stopCh)
	}
}

func (c *cslCollector) collectContainerCSL() {
	klog.V(6).Infof("start collectContainerCSL")
	timeWindow := time.Now()
	containerCgroupNames := []string{}
	containerInfos := map[string]metrics.CSLContainerInfo{}
	podMetas := c.statesInformer.GetAllPods()
	for _, meta := range podMetas {
		pod := meta.Pod
		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerDir, err := system.CgroupPathFormatter.ContainerDirFn(containerStatus.ContainerID)
			if err != nil {
				klog.Errorf("failed to get cgroup name of container %v from pod %v: %v", containerStatus.Name, pod.Name, err)
				return
			}
			containerCgroupName := containerDir[:len(containerDir)-1]
			containerCgroupNames = append(containerCgroupNames, containerCgroupName)
			containerInfos[containerCgroupName] = metrics.CSLContainerInfo{
				ContainerID:   containerStatus.ContainerID,
				ContainerName: containerStatus.Name,
				PodUID:        string(pod.UID),
				PodName:       pod.Name,
				PodNamespace:  pod.Namespace,
			}
		}
	}
	// todo: the usage of eBPF program to be discussed
	// option 1: load eBPF when koordetector starts, need to reset BPF maps in case of data overflow
	// option 2: keep current way, the whole eBPF program and related maps do load/unload every time window
	bpfSupported, err := supportedIfFileExists("/sys/kernel/btf/vmlinux")
	if err != nil {
		klog.Errorf("error checking vmlinux file: %v", err)
		return
	}
	prog, err := csl.NewCSLeBPFProg(bpfSupported)
	if err != nil {
		klog.Errorf("failed to load eBPF program : %v", err)
		return
	}
	defer func(prog *csl.ProgObjects) {
		err := prog.DestroyEBPFProg()
		if err != nil {
			klog.Fatalf("Destroy eBPF prog err: %v", err)
		}
	}(prog)

	time.Sleep(c.collectTimeWindowDuration)

	cgroupAvgLatency, err := prog.GetCgroupScheduleLatencyAvg(containerCgroupNames)
	if err != nil {
		klog.Errorf("failed to get container cpu schedule latency : %v", err)
		return
	}

	// todo: store csl into tsdb
	for cgroupName, latency := range cgroupAvgLatency {
		if containerInfo, ok := containerInfos[cgroupName]; ok {
			metrics.RecordContainerCSL(containerInfo, latency)
		} else {
			klog.Errorf("csl container info lost for cgroupName: %v, latency: %v", cgroupName, latency)
		}
	}
	c.started.Store(true)
	klog.V(5).Infof("collectContainerCSL for time window %s finished at %s, container num %d",
		timeWindow, time.Now(), len(containerCgroupNames))
	return
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func supportedIfFileExists(vmlinuxFilePath string) (bool, error) {
	exists, err := pathExists(vmlinuxFilePath)
	if err != nil {
		return false, fmt.Errorf("cannot check if vmlinux file exists, err: %v", err)
	}
	if !exists {
		return false, nil
	}
	return true, nil
}
