package k8s

import (
	"context"
	"sync"

	nodev1 "k8s.io/api/node/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type runtimeName string

const (
	vmAnnotationKey = "name.vm.kcrow.io"

	kataName runtimeName = "kata"
)

type RuntimeItem struct {
	Ev Event
	No *nodev1.RuntimeClass
}

type RuntimeManage struct {
	ctx context.Context

	syncedFn func() bool

	proc []RuntimeRegister

	mu       sync.RWMutex
	runtimes map[string]runtimeName
}

func NewRuntimeManage(ctx context.Context, reader cache.Cache) *RuntimeManage {
	rm := &RuntimeManage{
		ctx:      ctx,
		runtimes: map[string]runtimeName{},
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
	klog.V(2).Infof("regist runtime process %v", fn.Name())
	rm.proc = append(rm.proc, fn)
}

func (rm *RuntimeManage) IsKata(name string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.runtimes[name] == kataName
}

func (rm *RuntimeManage) Isvm(name string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	_, ok := rm.runtimes[name]
	return ok
}

func (rm *RuntimeManage) OnAdd(obj interface{}, isInInitialList bool) {
	if rm.handler(obj, AddEvent) {
		for _, p := range rm.proc {
			p.RuntimeUpdate(&RuntimeItem{
				Ev: AddEvent,
				No: obj.(*nodev1.RuntimeClass),
			})
		}
	}
}

func (rm *RuntimeManage) OnUpdate(oldObj, newObj interface{}) {
	if rm.handler(newObj, UpdateEvent) {
		for _, p := range rm.proc {
			p.RuntimeUpdate(&RuntimeItem{
				Ev: UpdateEvent,
				No: newObj.(*nodev1.RuntimeClass),
			})
		}
	}
}

func (rm *RuntimeManage) OnDelete(obj interface{}) {
	if rm.handler(obj, DeleteEvent) {
		for _, p := range rm.proc {
			p.RuntimeUpdate(&RuntimeItem{
				Ev: DeleteEvent,
				No: obj.(*nodev1.RuntimeClass),
			})
		}
	}
}

func (rm *RuntimeManage) handler(obj interface{}, ev Event) bool {
	rmobj, ok := obj.(*nodev1.RuntimeClass)
	if !ok {
		klog.Errorf("obj is invalid type, not runtimeclass")
		return false
	}
	if rmobj.Annotations != nil {
		rm.mu.Lock()
		defer rm.mu.Unlock()
		if ev == DeleteEvent {
			delete(rm.runtimes, rmobj.Name)
			return true
		}

		v, ok := rmobj.Annotations[vmAnnotationKey]
		if ok {
			rm.runtimes[rmobj.Name] = ""
		}
		switch v {
		case string(kataName):
			rm.runtimes[rmobj.Name] = kataName
		}
	}
	return true
}
