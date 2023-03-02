package config

import (
	"flag"
	"strings"

	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/koordinator-sh/koordetector/pkg/features"
)

type Configuration struct {
	KubeRestConf       *rest.Config
	FeatureGates       map[string]bool
}

func NewConfiguration() *Configuration {
	return &Configuration{}
}

func (c *Configuration) InitFlags(fs *flag.FlagSet) {
	fs.Var(cliflag.NewMapStringBool(&c.FeatureGates), "feature-gates", "A set of key=value pairs that describe feature gates for alpha/experimental features. "+
		"Options are:\n"+strings.Join(features.DefaultKoordetectorFeatureGate.KnownFeatures(), "\n"))
}

func (c *Configuration) InitClient() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	cfg.UserAgent = "koordlet"
	c.KubeRestConf = cfg
	return nil
}