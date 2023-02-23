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

// InterferenceDetectionRule CRD defines detection algorithms related information.
// Users can implement an interference detection algorithm based on particular performance metrics, and configure
// arguments such as Threshold into InterferenceDetectionRule's Spec. It can also determine on which workloads the
// interference detection feature will take effect from a high level.

// InterferenceDetectionRuleSpec defines the desired state of InterferenceDetectionRule
type InterferenceDetectionRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of InterferenceDetectionRule. Edit interferencedetectionrule_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// InterferenceDetectionRuleStatus is where the interference manager calculates a workload's normal performance
// and saved into.
// It will also be distributed on single nodes and work with strategies to determine if a workload is being interfered.
// For example, interference manager collects a workload's CPIs in the last 24 hours and calculates the standard
// deviation and mean value as its InterferenceDetectionRuleStatus, if at some moment the CPI is severely deviated from
// it we can say it is being interfered.

// InterferenceDetectionRuleStatus defines the observed state of InterferenceDetectionRule
type InterferenceDetectionRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InterferenceDetectionRule is the Schema for the interferencedetectionrules API
type InterferenceDetectionRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InterferenceDetectionRuleSpec   `json:"spec,omitempty"`
	Status InterferenceDetectionRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InterferenceDetectionRuleList contains a list of InterferenceDetectionRule
type InterferenceDetectionRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InterferenceDetectionRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InterferenceDetectionRule{}, &InterferenceDetectionRuleList{})
}
