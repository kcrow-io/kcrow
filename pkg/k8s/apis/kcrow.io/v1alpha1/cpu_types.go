/*
Copyright 2023.

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

// CpuSpec defines the desired state of resource.
type CpuSpec struct {

	// cpuset configure, support *
	Cpu *LinuxCPU `json:"cpu,omitempty"`

	// +kubebuilder:validation:Optional
	NamespaceName []string `json:"namespaceName,omitempty"`

	// +kubebuilder:validation:Optional
	RuntimeClassName []string `json:"runtimeClassName,omitempty"`
}

// CpusetStatus defines the observed state of resource.
type CpuStatus struct {
	// +kubebuilder:validation:Optional
	Pods map[string]*LinuxCPU `json:"pods,omitempty"`
}

// +kubebuilder:object:root=true
// Cpu is the Schema for the Cpus API.
type Cpu struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CpuSpec   `json:"spec,omitempty"`
	Status CpuStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CpuList contains a list of Cpu.
type CpuList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Cpu `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CpuList{}, &Cpu{})
}
