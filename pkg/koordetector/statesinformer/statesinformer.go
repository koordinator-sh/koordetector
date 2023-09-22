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

package statesinformer

// todo:
// statesinformer is a mature pkg in koordinator, we re-implement it here simply because we plan to use tsdb
// instead of sqlite metriccache in koordetector. Anyway, koordinator is working on tsdb, so this pkg need to
// be refactored in the future.

import (
	"fmt"

	_ "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/clientset/versioned/scheme"
	"go.uber.org/atomic"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	HTTPScheme  = "http"
	HTTPSScheme = "https"
)

type StatesInformer interface {
	Run(stopCh <-chan struct{}) error
	HasSynced() bool

	GetNode() *corev1.Node

	GetAllPods() []*PodMeta

	RegisterCallbacks(objType RegisterType, name, description string, callbackFn UpdateCbFn)
}

type pluginName string

type pluginOption struct {
	config     *Config
	KubeClient clientset.Interface
	NodeName   string
}

type pluginState struct {
	// todo: use tsdb
	// metricCache     metriccache.MetricCache
	informerPlugins map[pluginName]informerPlugin
	callbackRunner  *callbackRunner
}

type statesInformer struct {
	// TODO refactor device as plugin
	config *Config
	// todo: use tsdb
	//metricsCache metriccache.MetricCache

	option  *pluginOption
	states  *pluginState
	started *atomic.Bool
}

type informerPlugin interface {
	Setup(ctx *pluginOption, state *pluginState)
	Start(stopCh <-chan struct{})
	HasSynced() bool
}

func NewStatesInformer(config *Config, kubeClient clientset.Interface, nodeName string) StatesInformer {
	opt := &pluginOption{
		config:     config,
		KubeClient: kubeClient,
		NodeName:   nodeName,
	}
	stat := &pluginState{
		informerPlugins: map[pluginName]informerPlugin{},
		callbackRunner:  NewCallbackRunner(),
	}
	s := &statesInformer{
		config: config,

		option:  opt,
		states:  stat,
		started: atomic.NewBool(false),
	}
	s.initInformerPlugins()
	return s
}

func (s *statesInformer) setupPlugins() {
	for name, plugin := range s.states.informerPlugins {
		plugin.Setup(s.option, s.states)
		klog.V(2).Infof("plugin %v has been setup", name)
	}
}

func (s *statesInformer) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	klog.V(2).Infof("setup statesInformer")

	klog.V(2).Infof("starting callback runner")
	s.states.callbackRunner.Setup(s)
	go s.states.callbackRunner.Start(stopCh)

	klog.V(2).Infof("starting informer plugins")
	s.setupPlugins()
	s.startPlugins(stopCh)

	// waiting for node synced.
	klog.V(2).Infof("waiting for informer syncing")
	waitInformersSynced := s.waitForSyncFunc()
	if !cache.WaitForCacheSync(stopCh, waitInformersSynced...) {
		return fmt.Errorf("timed out waiting for states informer caches to sync")
	}

	klog.Infof("start states informer successfully")
	s.started.Store(true)
	<-stopCh
	klog.Infof("shutting down states informer daemon")
	return nil
}

func (s *statesInformer) waitForSyncFunc() []cache.InformerSynced {
	waitInformersSynced := make([]cache.InformerSynced, 0, len(s.states.informerPlugins))
	for _, p := range s.states.informerPlugins {
		waitInformersSynced = append(waitInformersSynced, p.HasSynced)
	}
	return waitInformersSynced
}

func (s *statesInformer) startPlugins(stopCh <-chan struct{}) {
	for name, p := range s.states.informerPlugins {
		klog.V(4).Infof("starting informer plugin %v", name)
		go p.Start(stopCh)
	}
}

func (s *statesInformer) HasSynced() bool {
	for _, p := range s.states.informerPlugins {
		if !p.HasSynced() {
			return false
		}
	}
	return true
}

func (s *statesInformer) GetNode() *corev1.Node {
	nodeInformerIf := s.states.informerPlugins[nodeInformerName]
	nodeInformer, ok := nodeInformerIf.(*nodeInformer)
	if !ok {
		klog.Fatalf("node informer format error")
	}
	return nodeInformer.GetNode()
}

func (s *statesInformer) GetAllPods() []*PodMeta {
	podsInformerIf := s.states.informerPlugins[podsInformerName]
	podsInformer, ok := podsInformerIf.(*podsInformer)
	if !ok {
		klog.Fatalf("pods informer format error")
	}
	return podsInformer.GetAllPods()
}

func (s *statesInformer) RegisterCallbacks(rType RegisterType, name, description string, callbackFn UpdateCbFn) {
	s.states.callbackRunner.RegisterCallbacks(rType, name, description, callbackFn)
}
