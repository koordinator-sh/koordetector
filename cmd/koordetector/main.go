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

package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/koordinator-sh/koordetector/cmd/koordetector/options"
	"github.com/koordinator-sh/koordetector/pkg/features"
	"github.com/koordinator-sh/koordetector/pkg/koordetector"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/config"
)

func init() {}

func main() {
	cfg := config.NewConfiguration()
	cfg.InitFlags(flag.CommandLine)
	flag.Parse()

	go wait.Forever(klog.Flush, 5*time.Second)
	defer klog.Flush()

	if *options.EnablePprof {
		go func() {
			klog.V(4).Infof("Starting pprof on %v", *options.PprofAddr)
			if err := http.ListenAndServe(*options.PprofAddr, nil); err != nil {
				klog.Errorf("Unable to start pprof on %v, error: %v", *options.PprofAddr, err)
			}
		}()
	}

	if err := features.DefaultMutableKoordetectorFeatureGate.SetFromMap(cfg.FeatureGates); err != nil {
		klog.Fatalf("Unable to setup feature-gates: %v", err)
	}

	stopCtx := signals.SetupSignalHandler()

	// Get a config to talk to the apiserver
	klog.Info("Setting up client for koordetector")
	err := cfg.InitClient()
	if err != nil {
		klog.Error("Unable to setup client config: ", err)
		os.Exit(1)
	}

	d, err := koordetector.NewDaemon(cfg)
	if err != nil {
		klog.Error("Unable to setup koordetector daemon: ", err)
		os.Exit(1)
	}

	// Expose the Prometheus http endpoint
	go func() {
		klog.Infof("Starting prometheus server on %v", *options.ServerAddr)
		http.Handle("/metrics", promhttp.Handler())
		// http.HandleFunc("/healthz", d.HealthzHandler())
		klog.Fatalf("Prometheus monitoring failed: %v", http.ListenAndServe(*options.ServerAddr, nil))
	}()

	// Start the Cmd
	klog.Info("Starting the koordetector daemon")
	d.Run(stopCtx.Done())
}
