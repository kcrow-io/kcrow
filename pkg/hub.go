package pkg

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
	merr "github.com/kcrow-io/kcrow/pkg/errors"
	"github.com/kcrow-io/kcrow/pkg/oci"
	"github.com/kcrow-io/kcrow/pkg/util"
	"k8s.io/klog/v2"
)

const (
	name  = "HubController"
	index = "05"
)

type Hub struct {
	rcs []oci.Oci
	ctx context.Context

	sockpath string
}

func New(ctx context.Context, nripath string, ocis ...oci.Oci) (*Hub, error) {

	hub := &Hub{
		ctx:      ctx,
		sockpath: nripath,
	}

	for _, oc := range ocis {
		if oc != nil {
			klog.Infof("registe controller '%v'", oc.Name())
			hub.rcs = append(hub.rcs, oc)
		}
	}
	return hub, nil
}

func (h *Hub) Start() {
	go func() {
		util.TimeBackoff(func() error { //nolint
			st, err := newStub(h, h.sockpath)
			if err != nil {
				klog.Errorf("init stub failed: %v", err)
				return err
			}
			err = st.Run(h.ctx)
			select {
			case <-h.ctx.Done():
				klog.Warning("context cancle, server exit.")
				return nil
			default:
				st.Stop()
				klog.Errorf("server exist, errmsg: %v", err)
				return fmt.Errorf("server exist, errmsg: %v", err)
			}
		}, 0)
	}()
}

func (h *Hub) CreateContainer(ctx context.Context, p *api.PodSandbox, ct *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	adjust := &api.ContainerAdjustment{
		Linux: &api.LinuxContainerAdjustment{
			Resources: ct.Linux.Resources,
		},
	}
	var (
		err error
		it  = &oci.Item{
			Adjust: adjust,
			Ct:     ct,
			Sb:     p,
		}
	)

	for _, rc := range h.rcs {
		err := rc.Process(ctx, it)
		if err != nil {
			if errors.Is(err, &merr.K8sError{}) || errors.Is(err, &merr.InternalError{}) {
				klog.Warningf("ignore controller %s error: %v", rc.Name(), err)
				err = nil // nolint
			} else {
				klog.Errorf("controller %s failed: %v", rc.Name(), err)
				break
			}
		}
	}
	return adjust, nil, err
}

func (h *Hub) StartContainer(ctx context.Context, p *api.PodSandbox, ct *api.Container) error {
	for _, rc := range h.rcs {
		// 调用每个插件的 Start 方法
		if err := rc.Start(ctx, p, ct); err != nil {
			klog.Errorf("controller %s start failed: %v", rc.Name(), err)
			return err
		}
	}
	return nil
}

func newStub(rc any, nripath string) (stub.Stub, error) {
	return stub.New(rc,
		stub.WithSocketPath(nripath),
		stub.WithPluginIdx(index),
		stub.WithPluginName(name),
		stub.WithOnClose(func() {}),
	)
}
