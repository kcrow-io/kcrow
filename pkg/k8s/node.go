package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type NodeItem struct {
	Ev Event
	No *corev1.Node
}

type NodeManage struct {
	ctx context.Context

	reader cache.Cache

	syncedFn func() bool

	proc []NodeRegister
}

// only record current node
func NewNodeControl(ctx context.Context, reader cache.Cache) *NodeManage {
	no := &NodeManage{
		reader: reader,
		ctx:    ctx,
	}
	err := no.probe()
	if err != nil {
		panic(err)
	}
	return no
}

// priority
func (no *NodeManage) probe() error {
	var (
		ns = &corev1.Namespace{}
	)
	informer, err := no.reader.GetInformer(no.ctx, ns)
	if err != nil {
		return err
	}
	evHandler := toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			v, ok := obj.(*corev1.Node)
			if !ok {
				return false
			}
			if v.Name == nodeName {
				return true
			}

			return false
		},
		Handler: no,
	}

	hadsync, err := informer.AddEventHandler(evHandler)
	if err != nil {
		return err
	}
	no.syncedFn = func() bool {
		return hadsync.HasSynced()
	}
	return nil
}

func (no *NodeManage) Registe(fn NodeRegister) {
	klog.Infof("regist node process %v", fn.Name())
	no.proc = append(no.proc, fn)
}

func (no *NodeManage) OnAdd(obj interface{}, isInInitialList bool) {
	for _, p := range no.proc {
		p.NodeUpdate(&NodeItem{
			Ev: AddEvent,
			No: obj.(*corev1.Node),
		})
	}
}

func (no *NodeManage) OnUpdate(oldObj, newObj interface{}) {
	for _, p := range no.proc {
		p.NodeUpdate(&NodeItem{
			Ev: UpdateEvent,
			No: newObj.(*corev1.Node),
		})
	}
}

func (no *NodeManage) OnDelete(obj interface{}) {
	for _, p := range no.proc {
		p.NodeUpdate(&NodeItem{
			Ev: DeleteEvent,
			No: obj.(*corev1.Node),
		})
	}

}

func TransNode(in interface{}) (out interface{}, err error) {
	v, ok := in.(*corev1.Node)
	if ok {
		return &corev1.Node{
			TypeMeta:   v.TypeMeta,
			ObjectMeta: v.ObjectMeta,
			Spec:       *v.Spec.DeepCopy(),
		}, nil
	}
	return nil, fmt.Errorf("it is not node type")
}
