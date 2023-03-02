package koordetector

import (
	"fmt"
	"github.com/koordinator-sh/koordinator/pkg/koordlet/metricsadvisor"
	"github.com/koordinator-sh/koordinator/pkg/koordlet/statesinformer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"os"
	"time"

	topologyclientset "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/clientset/versioned"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"

	"github.com/koordinator-sh/koordetector/pkg/koordetector/config"
	clientsetbeta1 "github.com/koordinator-sh/koordinator/pkg/client/clientset/versioned"
	"github.com/koordinator-sh/koordinator/pkg/client/clientset/versioned/typed/scheduling/v1alpha1"
)

var (
	scheme = apiruntime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

type Daemon interface {
	Run(stopCh <-chan struct{})
}

type daemon struct {
	collector      metricsadvisor.MetricAdvisor
	statesInformer statesinformer.StatesInformer
}

func NewDaemon(config *config.Configuration) (Daemon, error) {
	// get node name
	nodeName := os.Getenv("NODE_NAME")
	if len(nodeName) == 0 {
		return nil, fmt.Errorf("failed to new daemon: NODE_NAME env is empty")
	}
	klog.Infof("NODE_NAME is %v,start time %v", nodeName, float64(time.Now().Unix()))
	// metrics.RecordKoordletStartTime(nodeName, float64(time.Now().Unix()))

	kubeClient := clientset.NewForConfigOrDie(config.KubeRestConf)
	crdClient := clientsetbeta1.NewForConfigOrDie(config.KubeRestConf)
	topologyClient := topologyclientset.NewForConfigOrDie(config.KubeRestConf)
	schedulingClient := v1alpha1.NewForConfigOrDie(config.KubeRestConf)

	// only sync pod info
	statesInformerConf := statesinformer.NewDefaultConfig()
	statesInformer := statesinformer.NewStatesInformer(statesInformerConf, kubeClient, crdClient, topologyClient, metricCache, nodeName, schedulingClient)

	// add metric collector

	d := &daemon{
		statesInformer: statesInformer,
	}
	return d, nil
}

func (d *daemon) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	klog.Infof("Starting daemon")

	klog.Info("Start daemon successfully")
	<-stopCh
	klog.Info("Shutting down daemon")
}