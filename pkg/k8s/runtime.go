package k8s

import (
	"context"
	"strings"
	"sync"

	nodev1 "k8s.io/api/node/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

const (
	vmAnnotationKey = "vm.kcrow.io"
)

type RuntimeItem struct {
	Ev Event
	No *nodev1.RuntimeClass
}

type RuntimeManage struct {
	ctx context.Context

	syncedFn func() bool

	proc []RuntimeRegister

	mu        sync.RWMutex
	vmruntime map[string]bool
}

func NewRuntimeManage(ctx context.Context, reader cache.Cache) *RuntimeManage {
	rm := &RuntimeManage{
		vmruntime: map[string]bool{},
	}
	err := rm.probe(reader)
	if err != nil {
		panic(err)
	}
	return rm
}

// priority
func (rm *RuntimeManage) probe(reader cache.Cache) error {
	var (
		ns = &nodev1.RuntimeClass{}
	)
	informer, err := reader.GetInformer(rm.ctx, ns)
	if err != nil {
		return err
	}
	evHandler := toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			_, ok := obj.(*nodev1.RuntimeClass)
			if !ok {
				return false
			}
			return false
		},
		Handler: rm,
	}

	hadsync, err := informer.AddEventHandler(evHandler)
	if err != nil {
		return err
	}
	rm.syncedFn = func() bool {
		return hadsync.HasSynced()
	}
	return nil
}

// regist process function, call when sync
func (rm *RuntimeManage) Registe(fn RuntimeRegister) {
	klog.Infof("regist runtime process %v", fn.Name())
	rm.proc = append(rm.proc, fn)
}

// regist process function, call when sync
func (rm *RuntimeManage) Isvm(name string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.vmruntime[name]
}

func (rm *RuntimeManage) OnAdd(obj interface{}, isInInitialList bool) {
	rmobj, ok := obj.(*nodev1.RuntimeClass)
	if !ok {
		return
	}
	var isvm bool
	for k := range rmobj.Annotations {
		if strings.ToLower(k) == string(vmAnnotationKey) {
			isvm = true
		}
	}
	for _, p := range rm.proc {
		p.RuntimeUpdate(&RuntimeItem{
			Ev: AddEvent,
			No: obj.(*nodev1.RuntimeClass),
		})
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.vmruntime[rmobj.Name] = isvm
}

func (rm *RuntimeManage) OnUpdate(oldObj, newObj interface{}) {
	rmobj, ok := newObj.(*nodev1.RuntimeClass)
	if !ok {
		return
	}
	var isvm bool
	for k := range rmobj.Annotations {
		if strings.ToLower(k) == string(vmAnnotationKey) {
			isvm = true
		}
	}
	for _, p := range rm.proc {
		p.RuntimeUpdate(&RuntimeItem{
			Ev: UpdateEvent,
			No: newObj.(*nodev1.RuntimeClass),
		})
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.vmruntime[rmobj.Name] = isvm
}

func (rm *RuntimeManage) OnDelete(obj interface{}) {
	for _, p := range rm.proc {
		p.RuntimeUpdate(&RuntimeItem{
			Ev: DeleteEvent,
			No: obj.(*nodev1.RuntimeClass),
		})
	}
}
