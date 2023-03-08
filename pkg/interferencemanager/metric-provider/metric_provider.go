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

package metric_provider

import (
	"fmt"

	"github.com/koordinator-sh/koordetector/pkg/interferencemanager/metric-provider/common"
	"github.com/koordinator-sh/koordetector/pkg/interferencemanager/metric-provider/config"
	"github.com/koordinator-sh/koordetector/pkg/interferencemanager/metric-provider/prometheus"
)

type MetricProvider interface {
	GetCPI(options common.MetricQueryOptions, labelFunc common.MakeLabelsFunc) ([]*common.Metric, error)
}

func NewMetricsProvider(config config.MetricProviderConfig) (MetricProvider, error) {
	switch config.ProviderType {
	case common.PrometheusProvider:
		provider, err := prometheus.NewPrometheusProvider(config.PromConf)
		if err != nil {
			return nil, err
		}
		return provider, nil
	}
	return nil, fmt.Errorf("metric provider does not support type: %v", config.ProviderType)
}
