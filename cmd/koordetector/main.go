package koordetector

import (
	"flag"
	"github.com/koordinator-sh/koordetector/cmd/koordetector/options"
	"github.com/koordinator-sh/koordetector/pkg/features"
	"github.com/koordinator-sh/koordetector/pkg/koordetector"
	"github.com/koordinator-sh/koordetector/pkg/koordetector/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"time"
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
	klog.Info("Setting up client for koordlet")
	err := cfg.InitClient()
	if err != nil {
		klog.Error("Unable to setup client config: ", err)
		os.Exit(1)
	}

	d, err := koordetector.NewDaemon(cfg)
	if err != nil {
		klog.Error("Unable to setup koordlet daemon: ", err)
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
	klog.Info("Starting the koordlet daemon")
	d.Run(stopCtx.Done())
}
