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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ContainerCSL = NewGCGaugeVec("container_csl", prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KoordetectorSubsystem,
		Name:      "container_csl",
		Help:      "Container cpu schedule latecy collected by koordetector",
	}, []string{NodeKey, ContainerID, ContainerName, PodUID, PodName, PodNamespace}))

	CSLCollectors = []prometheus.Collector{
		ContainerCSL.GetGaugeVec(),
	}
)

type CSLContainerInfo struct {
	ContainerID   string
	ContainerName string
	PodUID        string
	PodName       string
	PodNamespace  string
}

func RecordContainerCSL(containerInfos CSLContainerInfo, latency float64) {
	labels := genNodeLabels()
	if labels == nil {
		return
	}
	labels[ContainerID] = containerInfos.ContainerID
	labels[ContainerName] = containerInfos.ContainerName
	labels[PodUID] = containerInfos.PodUID
	labels[PodName] = containerInfos.PodName
	labels[PodNamespace] = containerInfos.PodNamespace
	ContainerCSL.WithSet(labels, latency)
}
