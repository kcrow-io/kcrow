package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type PodManage struct {
	ctx context.Context

	reader cache.Cache

	HadSynced func() bool
}

func NewPodControl(ctx context.Context, reader cache.Cache) *PodManage {
	nr := &PodManage{
		ctx:    ctx,
		reader: reader,
	}
	return nr
}

func (nr *PodManage) Pod(nsname types.NamespacedName) (*corev1.Pod, error) {
	var (
		po = &corev1.Pod{}
	)
	err := nr.reader.Get(nr.ctx, nsname, po)
	if err != nil {
		return nil, err
	}
	return po, nil
}

func TransPod(in interface{}) (out interface{}, err error) {
	v, ok := in.(*corev1.Pod)
	if ok {
		return &corev1.Pod{
			TypeMeta:   v.TypeMeta,
			ObjectMeta: v.ObjectMeta,
		}, nil
	}
	return nil, fmt.Errorf("not pod type")
}
