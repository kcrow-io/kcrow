package pkg

import (
	"context"

	"github.com/containerd/nri/pkg/stub"
	"github.com/yylt/kcrow/pkg/resource"
	"github.com/yylt/kcrow/pkg/util"
)

const (
	name  = "ResourceController"
	index = "0"
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
		util.TimeBackoff(func() error {
			return h.s.Start(h.ctx)
		}, 0)
	}()
}
