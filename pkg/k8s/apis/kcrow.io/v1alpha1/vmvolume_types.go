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

type voltype string

const (
	NfsVol    voltype = "nfs"
	RbdVol    voltype = "rbd"
	DeviceVol voltype = "device"
)

// translate mount to device and hooks
// VmVolumeSpec defines the desired state of resource.
type VmVolumeSpec struct {

	// storageclass name and voltype
	// voltype will be used in hook command, like mount
	VolumeMap map[string]voltype `json:"volumeMap,omitempty"`

	// +kubebuilder:validation:Optional
	RuntimeClassName []string `json:"runtimeClassName,omitempty"`
}

// VmVolumeStatus defines the observed state of resource.
type VmVolumeStatus struct {
	// +kubebuilder:validation:Optional
	Volume []PodVmVolume `json:"volume,omitempty"`
}

// +kubebuilder:object:root=true
// VmVolume is the Schema for the VmVolumes API.
type VmVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VmVolumeSpec   `json:"spec,omitempty"`
	Status VmVolumeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VmVolumeList contains a list of VmVolume.
type VmVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VmVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VmVolumeList{}, &VmVolume{})
}
