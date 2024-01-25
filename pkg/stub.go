package pkg

import (
	"github.com/containerd/nri/pkg/api"
)

type Handler interface {
	Name() string
}

type Createhandler interface {
	CreateContainer(*api.PodSandbox, *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error)
}

type Updatehandler interface {
	UpdateContainer(*api.PodSandbox, *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error)
}

type Stophandler interface {
	StopContainer(*api.PodSandbox, *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error)
}

func RegistHandler(h Handler) {

}
