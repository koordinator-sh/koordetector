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

package common

import (
	prommodel "github.com/prometheus/common/model"
)

type MetricQueryOptions struct {
	MetricName   string
	FilterLabels map[string]string

	PromSumByLabels []string
}

type Metric struct {
	Labels map[string]string
	Value  float64
}

type ProviderType string

const (
	PrometheusProvider ProviderType = "prometheus_provider"
)

const (
	ContainerID   string = "container_id"
	ContainerName string = "container_name"
	Node          string = "node"
	PodName       string = "pod_name"
	PodNamespace  string = "pod_namespace"
	PodUID        string = "pod_uid"

	CPIField string = "cpi_field"
)

const (
	KoordletContainerCPI string = "koordlet_container_cpi"
	KoordletPodCPI       string = "koordlet_pod_cpi"

	Cycles       string = "cycles"
	Instructions string = "instructions"
)

type MakeLabelsFunc func(metric prommodel.Metric) (map[string]string, error)

func MakeContainerCPILabels(metric prommodel.Metric) (map[string]string, error) {
	labels := map[string]string{
		ContainerID:   string(metric["container_id"]),
		ContainerName: string(metric["container_name"]),
		PodUID:        string(metric["pod_uid"]),
		PodNamespace:  string(metric["pod_namespace"]),
		PodName:       string(metric["pod_name"]),
		Node:          string(metric["node"]),
	}
	return labels, nil
}

func MakePodCPILabels(metric prommodel.Metric) (map[string]string, error) {
	labels := map[string]string{
		PodUID:       string(metric["pod_uid"]),
		PodNamespace: string(metric["pod_namespace"]),
		PodName:      string(metric["pod_name"]),
		Node:         string(metric["node"]),
	}
	return labels, nil
}
