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

package prometheus

import (
	"context"
	"fmt"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"

	"github.com/koordinator-sh/koordetector/pkg/interferencemanager/metric-provider/common"
	"github.com/koordinator-sh/koordetector/pkg/interferencemanager/metric-provider/config"
)

const (
	SumBy string = "sum by"
)

type prometheusProvider struct {
	prometheusClient prometheusv1.API
	config           config.PrometheusProviderConfig
	queryTimeout     time.Duration
}

// NewPrometheusProvider contructs a metric provider that gets data from Prometheus.
func NewPrometheusProvider(config config.PrometheusProviderConfig) (*prometheusProvider, error) {
	promClient, err := promapi.NewClient(promapi.Config{
		Address: config.Address,
	})
	if err != nil {
		return &prometheusProvider{}, err
	}
	return &prometheusProvider{
		prometheusClient: prometheusv1.NewAPI(promClient),
		config:           config,
		queryTimeout:     config.QueryTimeout,
	}, nil
}

func (p *prometheusProvider) GetCPI(options common.MetricQueryOptions, labelFunc common.MakeLabelsFunc) ([]*common.Metric, error) {
	query, err := MakeQueryCPIString(options)
	if err != nil {
		return nil, err
	}
	var result []*common.Metric
	promResult, err := p.query(query)
	if err != nil {
		return nil, err
	}
	for _, metric := range promResult {
		labels, err := labelFunc(metric.Metric)
		if err != nil {
			return nil, err
		}
		result = append(result, &common.Metric{
			Labels: labels,
			Value:  float64(metric.Value),
		})
	}
	return result, nil
}

// query delicate prometheus query api.
func (p *prometheusProvider) query(query string) (prommodel.Vector, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.queryTimeout)
	defer cancel()

	result, _, err := p.prometheusClient.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("cannot get metrics for query %v: %v", query, err)
	}

	vector, ok := result.(prommodel.Vector)
	if !ok {
		return nil, fmt.Errorf("expected query to return a vector; got result type %T", result)
	}

	return vector, nil
}

func NewDefaultCPISumByLabels(metricName string) ([]string, error) {
	switch metricName {
	case common.KoordletContainerCPI:
		return []string{
			common.ContainerID,
			common.ContainerName,
			common.PodUID,
			common.PodNamespace,
			common.PodName,
			common.Node,
		}, nil
	case common.KoordletPodCPI:
		return []string{
			common.PodUID,
			common.PodNamespace,
			common.PodName,
			common.Node,
		}, nil
	}
	return nil, fmt.Errorf("metric name %v not supported", metricName)
}

// MakeQueryCPIString constructs PromQL style query string based on @options.
//
// @return
// "sum by (container_id, container_name, pod_uid, pod_namespace, pod_name, node) " +
// "(koordlet_container_cpi{cpi_field=\"cycles\"}) / " +
// "sum by (container_id, container_name, pod_uid, pod_namespace, pod_name, node) " +
// "(koordlet_container_cpi{cpi_field=\"instructions\"})"
//
// where
// (container_id, container_name, pod_uid, pod_namespace, pod_name, node) is the PromSumByLabels slice
// {cpi_field=\"cycles\"} is the FilterLabels map
// koordlet_container_cpi is options.MetricName
func MakeQueryCPIString(options common.MetricQueryOptions) (string, error) {
	if options.PromSumByLabels == nil {
		labels, err := NewDefaultCPISumByLabels(options.MetricName)
		if err != nil {
			return "", err
		}
		options.PromSumByLabels = labels
	}
	if options.FilterLabels == nil {
		options.FilterLabels = map[string]string{}
	}
	options.FilterLabels[common.CPIField] = common.Cycles
	queryCycles := makeSumByString(options.PromSumByLabels) + makeMetricFilterString(common.KoordletContainerCPI, options.FilterLabels)
	options.FilterLabels[common.CPIField] = common.Instructions
	queryInstructions := makeSumByString(options.PromSumByLabels) + makeMetricFilterString(common.KoordletContainerCPI, options.FilterLabels)
	return queryCycles + "/" + queryInstructions, nil
}

func makeSumByString(labels []string) string {
	labelsString := strings.Join(labels, ",")
	return fmt.Sprintf("%v(%v)", SumBy, labelsString)
}

func makeMetricFilterString(metricName string, filters map[string]string) string {
	var filterSlice []string
	for label, value := range filters {
		filterSlice = append(filterSlice, fmt.Sprintf("%v=\"%v\"", label, value))
	}
	filterString := strings.Join(filterSlice, ",")
	return fmt.Sprintf("(%v{%v})", metricName, filterString)
}
