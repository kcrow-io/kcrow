package pkg

import (
	"context"
	"fmt"

	"github.com/containerd/nri/pkg/stub"
	"github.com/yylt/kcrow/pkg/resource"
	"github.com/yylt/kcrow/pkg/util"
	"k8s.io/klog/v2"
)

const (
	name  = "ResourceController"
	index = "05"
)

type Hub struct {
	rc  *resource.ResManage
	ctx context.Context
}

func New(rc *resource.ResManage, ctx context.Context) (*Hub, error) {
	_, err := newStub(rc)
	if err != nil {
		return nil, err
	}
	return &Hub{
		rc:  rc,
		ctx: ctx,
	}, nil
}

func (h *Hub) Start() {
	go func() {
		util.TimeBackoff(func() error { //nolint
			st, err := newStub(h.rc)
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
				klog.Warningf("server exit: %v", err)
				st.Stop()
				return fmt.Errorf("server exist, errmsg: %v", err)
			}
		}, 0)
	}()
}

func newStub(rc any) (stub.Stub, error) {
	return stub.New(rc,
		stub.WithPluginIdx(index),
		stub.WithPluginName(name),
		stub.WithOnClose(func() {}),
	)
}
