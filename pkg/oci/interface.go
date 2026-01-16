package oci

import (
	"context"

	"github.com/containerd/nri/pkg/api"
	"k8s.io/apimachinery/pkg/types"
)

const (
	PodName      = "io.kubernetes.pod.name"
	PodNamespace = "io.kubernetes.pod.namespace"
)

type Item struct {
	Sb     *api.PodSandbox
	Ct     *api.Container
	Adjust *api.ContainerAdjustment
}

type Oci interface {
	Name() string
	Process(context.Context, *Item) error
	Start(ctx context.Context, pod *api.PodSandbox, container *api.Container) error
}

func GetPodInfo(ct *api.Container) types.NamespacedName {
	if ct == nil || ct.Labels == nil {
		return types.NamespacedName{}
	}
	return types.NamespacedName{
		Namespace: GetNamespace(ct),
		Name:      GetName(ct),
	}
}

func GetNamespace(ct *api.Container) string {
	if ct == nil || ct.Labels == nil {
		return ""
	}
	return ct.Labels[PodNamespace]
}

func GetName(ct *api.Container) string {
	if ct == nil || ct.Labels == nil {
		return ""
	}
	return ct.Labels[PodName]
}
