package k8s

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type NsItem struct {
	Ev Event
	Ns *corev1.Namespace
}

type NsManage struct {
	ctx context.Context

	reader cache.Cache

	HadSynced func() bool

	proc []NamespaceRegister
}

func NewNsControl(ctx context.Context, reader cache.Cache) *NsManage {
	nr := &NsManage{
		ctx:    ctx,
		reader: reader,
	}
	err := nr.probe()
	if err != nil {
		panic(err)
	}
	return nr
}

// priority
func (nr *NsManage) probe() error {
	var (
		ns = &corev1.Namespace{}
	)
	informer, err := nr.reader.GetInformer(nr.ctx, ns)
	if err != nil {
		return err
	}
	evHandler := toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			_, ok := obj.(*corev1.Namespace)
			return ok
		},
		Handler: nr,
	}

	hadsync, err := informer.AddEventHandler(evHandler)
	if err != nil {
		return err
	}
	nr.HadSynced = func() bool {
		return hadsync.HasSynced()
	}
	return nil
}

// regist process function, call when sync
func (no *NsManage) Registe(fn NamespaceRegister) {
	klog.Infof("regist namespace callback %v", fn.Name())
	no.proc = append(no.proc, fn)
}

func (nr *NsManage) OnAdd(obj interface{}, isInInitialList bool) {
	for _, p := range nr.proc {
		p.NamespaceUpdate(&NsItem{
			Ev: AddEvent,
			Ns: obj.(*corev1.Namespace),
		})
	}
}

func (nr *NsManage) OnUpdate(oldObj, newObj interface{}) {
	oldNs := oldObj.(*corev1.Namespace)
	newNs := newObj.(*corev1.Namespace)
	if reflect.DeepEqual(oldNs.ObjectMeta.Annotations, newNs.ObjectMeta.Annotations) {
		return
	}
	for _, p := range nr.proc {
		p.NamespaceUpdate(&NsItem{
			Ev: UpdateEvent,
			Ns: newNs,
		})
	}
}

func (nr *NsManage) OnDelete(obj interface{}) {
	for _, p := range nr.proc {
		p.NamespaceUpdate(&NsItem{
			Ev: DeleteEvent,
			Ns: obj.(*corev1.Namespace),
		})
	}
}
