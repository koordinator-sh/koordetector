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

package features

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"
)

const (
	// owner: @songtao98 @zwzhang0107
	// alpha: v1.0
	//
	// CSLCollector enables cpu schedule latency collector feature of koordetector.
	CSLCollector featuregate.Feature = "CSLCollector"
)

func init() {
	runtime.Must(DefaultMutableKoordetectorFeatureGate.Add(defaultKoordetectorFeatureGates))
}

var (
	DefaultMutableKoordetectorFeatureGate featuregate.MutableFeatureGate = featuregate.NewFeatureGate()
	DefaultKoordetectorFeatureGate        featuregate.FeatureGate        = DefaultMutableKoordetectorFeatureGate

	defaultKoordetectorFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
		CSLCollector: {Default: false, PreRelease: featuregate.Alpha},
	}
)
