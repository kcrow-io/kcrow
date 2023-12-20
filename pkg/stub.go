package pkg

import (
	"github.com/containerd/nri/pkg/api"
)

type Handler interface {
	Name() string
	CreateContainer(*api.PodSandbox, *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error)
	UpdateContainer(*api.PodSandbox, *api.Container, *api.LinuxResources) ([]*api.ContainerUpdate, error)
	StopContainer(*api.PodSandbox, *api.Container) ([]*api.ContainerUpdate, error)
}

func RegistHandler(h Handler) {

}
