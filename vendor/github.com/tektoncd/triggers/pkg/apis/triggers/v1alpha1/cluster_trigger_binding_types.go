/*
Copyright 2019 The Tekton Authors

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
	"knative.dev/pkg/apis"
)

// Check that ClusterTriggerBinding may be validated and defaulted.
var _ apis.Validatable = (*ClusterTriggerBinding)(nil)
var _ apis.Defaultable = (*ClusterTriggerBinding)(nil)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// ClusterTriggerBinding is a TriggerBinding with a cluster scope.
// ClusterTriggerBindings are used to represent TriggerBindings that
// should be publicly addressable from any namespace in the cluster.
type ClusterTriggerBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the ClusterTriggerBinding from the client
	// +optional
	Spec TriggerBindingSpec `json:"spec,omitempty"`

	// +optional
	Status TriggerBindingStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterTriggerBindingList contains a list of ClusterTriggerBinding
type ClusterTriggerBindingList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterTriggerBinding `json:"items"`
}

func (ctb *ClusterTriggerBinding) TriggerBindingSpec() TriggerBindingSpec {
	return ctb.Spec
}

func (ctb *ClusterTriggerBinding) TriggerBindingMetadata() metav1.ObjectMeta {
	return ctb.ObjectMeta
}

func (ctb *ClusterTriggerBinding) Copy() TriggerBindingInterface {
	return ctb.DeepCopy()
}
