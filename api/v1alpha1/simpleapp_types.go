/*
Copyright 2023 vpeti12.

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

// SimpleAppSpec defines the desired state of SimpleApp
type SimpleAppSpec struct {
	// number of replicas
	Replicas *int32 `json:"replicas,omitempty"`
	// host where the application is accessible (ie. in the browser)
	Host string `json:"host,omitempty"`
	// container image (and tag) (eg. "nginx:latest")
	Image string `json:"image,omitempty"`
}

// SimpleAppStatus defines the observed state of SimpleApp
type SimpleAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SimpleApp is the Schema for the simpleapps API
type SimpleApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimpleAppSpec   `json:"spec,omitempty"`
	Status SimpleAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SimpleAppList contains a list of SimpleApp
type SimpleAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SimpleApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SimpleApp{}, &SimpleAppList{})
}
