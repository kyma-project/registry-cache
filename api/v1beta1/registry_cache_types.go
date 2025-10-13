/*
Copyright 2025.

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

// +kubebuilder:object:generate=true
// +groupName=operator.kyma-project.io
//

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories={kyma-modules,kyma-registry-cache}
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=".status.state", description="State of Registry Cache"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age of the resource"

// RegistryCache is the Schema for the registrycache API
type RegistryCache struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistryCacheSpec   `json:"spec,omitempty"`
	Status RegistryCacheStatus `json:"status,omitempty"`
}

// RegistryCacheSpec defines the desired state of RegistryCache
type RegistryCacheSpec struct{}

// Valid RegistryCache States.
const (
	// StateReady signifies RegistryCache is ready and has been installed successfully.
	StateReady State = "Ready"

	// StateProcessing signifies RegistryCache is reconciling and is in the process of installation.
	// Processing can also signal that the Installation previously encountered an error and is now recovering.
	StateProcessing State = "Processing"

	// StateWarning signifies a warning for RegistryCache. This signifies that the Installation
	// process encountered a problem.
	StateWarning State = "Warning"

	// StateError signifies an error for RegistryCache. This signifies that the Installation
	// process encountered an error.
	// Contrary to Processing, it can be expected that this state should change on the next retry.
	StateError State = "Error"

	// StateDeleting signifies RegistryCache is being deleted. This is the state that is used
	// when a deletionTimestamp was detected and Finalizers are picked up.
	StateDeleting State = "Deleting"
)

// +k8s:deepcopy-gen=true

// RegistryCacheStatus defines the observed state of RegistryCache
type RegistryCacheStatus struct {
	// State signifies current state of Module CR.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting", "Warning", or empty).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning;""
	State State `json:",inline"`

	// Conditions contain a set of conditionals to determine the State of Status.
	// If all Conditions are met, State is expected to be in StateReady.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

func (s *RegistryCacheStatus) WithState(state State) *RegistryCacheStatus {
	s.State = state
	return s
}

// +kubebuilder:object:root=true
// RegistryCacheList contains a list of RegistryCache
type RegistryCacheList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RegistryCache `json:"items"`
}
