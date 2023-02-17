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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InterferenceMetricCheckpoint CRD is used to persistently stores metrics. Interference manager collects interference
// metrics from datasources(currently Prometheus) and aggregate those metrics into Histograms. This process is done in
// memory which is not safe enough. By using this CRD these metrics can be restored to initialize the model after
// restart.

// InterferenceMetricCheckpointSpec defines the desired state of InterferenceMetricCheckpoint
type InterferenceMetricCheckpointSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of InterferenceMetricCheckpoint. Edit interferencemetriccheckpoint_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// InterferenceMetricCheckpointStatus defines the observed state of InterferenceMetricCheckpoint
type InterferenceMetricCheckpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InterferenceMetricCheckpoint is the Schema for the interferencemetriccheckpoints API
type InterferenceMetricCheckpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InterferenceMetricCheckpointSpec   `json:"spec,omitempty"`
	Status InterferenceMetricCheckpointStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InterferenceMetricCheckpointList contains a list of InterferenceMetricCheckpoint
type InterferenceMetricCheckpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InterferenceMetricCheckpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InterferenceMetricCheckpoint{}, &InterferenceMetricCheckpointList{})
}
