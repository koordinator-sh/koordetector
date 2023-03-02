package features

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"
)

func init() {
	runtime.Must(DefaultMutableKoordetectorFeatureGate.Add(defaultKoordetectorFeatureGates))
}

var (
	DefaultMutableKoordetectorFeatureGate featuregate.MutableFeatureGate = featuregate.NewFeatureGate()
	DefaultKoordetectorFeatureGate        featuregate.FeatureGate        = DefaultMutableKoordetectorFeatureGate

	defaultKoordetectorFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
		
	}
)