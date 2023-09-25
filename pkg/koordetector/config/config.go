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

package config

import (
	"flag"
	"strings"

	"github.com/koordinator-sh/koordinator/pkg/koordlet/util/system"
	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/koordinator-sh/koordetector/pkg/features"
	maframework "github.com/koordinator-sh/koordetector/pkg/koordetector/metricsadvisor/framework"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/statesinformer"
)

type Configuration struct {
	KubeRestConf       *rest.Config
	FeatureGates       map[string]bool
	StatesInformerConf *statesinformer.Config
	CollectorConf      *maframework.Config
}

func NewConfiguration() *Configuration {
	return &Configuration{
		StatesInformerConf: statesinformer.NewDefaultConfig(),
		CollectorConf:      maframework.NewDefaultConfig(),
	}
}

func (c *Configuration) InitFlags(fs *flag.FlagSet) {
	fs.Var(cliflag.NewMapStringBool(&c.FeatureGates), "feature-gates", "A set of key=value pairs that describe feature gates for alpha/experimental features. "+
		"Options are:\n"+strings.Join(features.DefaultKoordetectorFeatureGate.KnownFeatures(), "\n"))
	system.Conf.InitFlags(fs)
	c.StatesInformerConf.InitFlags(fs)
	c.CollectorConf.InitFlags(fs)
}

func (c *Configuration) InitClient() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	cfg.UserAgent = "koordetector"
	c.KubeRestConf = cfg
	return nil
}
