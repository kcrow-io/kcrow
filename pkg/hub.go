package pkg

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/cluster"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
	"github.com/yylt/kcrow/pkg/resource"
)

const (
	name  = "ResourceController"
	index = "0"
)

type ResourceController struct {
	rsc *resource.ResManage

	ctx context.Context

	s stub.Stub
}

func NewResourceController(mgr cluster.Cluster, ctx context.Context) (*ResourceController, error) {
	var err error

	rc := &ResourceController{
		ctx: ctx,
	}
	node := resource.NewNodeControl(ctx, mgr.GetCache())
	namespace := resource.NewNsControl(ctx, mgr.GetCache())
	rc.rsc = resource.New(namespace, node)

	rc.s, err = stub.New(rc,
		stub.WithPluginIdx(index),
		stub.WithPluginName(name),
	)
	if err != nil {
		return nil, err
	}
	return rc, rc.probe()
}

func (h *ResourceController) probe() error {

}

func (h *ResourceController) Start() error {

}

func (h *ResourceController) CreateContainer(context.Context, *api.PodSandbox, *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {

}
