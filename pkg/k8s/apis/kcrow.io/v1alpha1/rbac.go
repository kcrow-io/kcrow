// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

// +kubebuilder:rbac:groups=kcrow.io,resources=cpu,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources="",verbs=get;list;watch;update

package v1alpha1
