package pkg

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Hub struct {
	client.Client
	ctx context.Context
}

func New(mgr ctrl.Manager, ctx context.Context) (*Hub, error) {
	hub := &Hub{
		Client: mgr.GetClient(),
		ctx:    ctx,
	}
	return hub, hub.probe()
}

func (h *Hub) probe() error {

}
