package pkg

import (
	"context"

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
	s   stub.Stub
	ctx context.Context
}

func New(rc *resource.ResManage, ctx context.Context) (*Hub, error) {
	s, err := stub.New(rc,
		stub.WithPluginIdx(index),
		stub.WithPluginName(name),
	)
	if err != nil {
		return nil, err
	}
	return &Hub{
		s:   s,
		ctx: ctx,
	}, nil
}

func (h *Hub) Start() {
	go func() {
		util.TimeBackoff(func() error { //nolint
			err := h.s.Run(h.ctx)
			if err != nil {
				klog.Errorf("nri hub start failed:%v", err)
				h.s.Stop()
			}
			select {
			case <-h.ctx.Done():
				klog.Warning("context cancle, nri hub exit.")
				return nil
			default:
				klog.Warning("server exit, msg: %v", err)
				return err
			}
		}, 0)
	}()
}
