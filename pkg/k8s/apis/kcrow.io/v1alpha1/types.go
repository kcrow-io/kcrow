// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

type LinuxCPU struct {
	// CPUs to use within the cpuset. Default is to use any CPU available.
	Cpus string `json:"cpus,omitempty"`
	// List of memory nodes in the cpuset. Default is to use any available memory node.
	Mems string `json:"mems,omitempty"`
	// cgroups are configured with minimum weight, 0: default behavior, 1: SCHED_IDLE.
	Idle *int64 `json:"idle,omitempty"`
}

type PodVmVolume struct {
	Pod    string `json:"pod,omitempty"`
	Volume string `json:"volume,omitempty"`
	Type   string `json:"type,omitempty"`
	Pv     string `json:"persistentVolume,omitempty"`
}
