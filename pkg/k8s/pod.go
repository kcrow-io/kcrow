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
	if reader == nil {
		panic(fmt.Errorf("reader cannot be nil"))
	}
	nr := &PodManage{
		ctx:    ctx,
		reader: reader,
	}
	return nr
}

func (nr *PodManage) Pod(nsname types.NamespacedName) (*corev1.Pod, error) {
	po := &corev1.Pod{}
	err := nr.reader.Get(nr.ctx, nsname, po)
	if err != nil {
		return nil, err
	}
	return po, nil
}

func TransPod(in interface{}) (out interface{}, err error) {
	v, ok := in.(*corev1.Pod)
	if ok {
		v.ManagedFields = nil
		v.Status = corev1.PodStatus{}
		return v, nil
	}
	return nil, fmt.Errorf("not pod type")
}
