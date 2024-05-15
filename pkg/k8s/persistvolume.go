package k8s

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type VolumeItem struct {
	Ev  Event
	Vol *corev1.PersistentVolume
}

type VolumeManage struct {
	ctx context.Context

	syncedFn func() bool

	mu sync.RWMutex

	// namespace/name (pv)
	specs map[string]*corev1.PersistentVolume

	proc []VolumeRegister
}

func NewVolumeManage(ctx context.Context, reader cache.Cache) *VolumeManage {
	vm := &VolumeManage{
		ctx:   ctx,
		specs: make(map[string]*corev1.PersistentVolume),
	}
	err := vm.probe(reader)
	if err != nil {
		panic(err)
	}
	return vm
}

func (vm *VolumeManage) probe(reader cache.Cache) error {
	var (
		ns = &corev1.PersistentVolume{}
	)
	infovmer, err := reader.GetInformer(vm.ctx, ns)
	if err != nil {
		return err
	}
	evHandler := toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			_, ok := obj.(*corev1.PersistentVolume)
			return ok
		},
		Handler: vm,
	}

	hadsync, err := infovmer.AddEventHandler(evHandler)
	if err != nil {
		return err
	}
	vm.syncedFn = func() bool {
		return hadsync.HasSynced()
	}
	return nil
}

// regist process function, call when sync
func (vm *VolumeManage) Registe(fn VolumeRegister) {
	klog.V(2).Infof("regist persistent volume callback %v", fn.Name())
	vm.proc = append(vm.proc, fn)
}

func (vm *VolumeManage) GetVolumeSpec(pvc types.NamespacedName) *corev1.PersistentVolume {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.specs[pvc.String()]
}

func (vm *VolumeManage) handler(obj interface{}, ev Event) bool {
	vol, ok := obj.(*corev1.PersistentVolume)
	if !ok {
		klog.Errorf("obj is invalid type, not persistentVolume")
		return false
	}
	if vol.Spec.ClaimRef != nil {
		name := types.NamespacedName{Namespace: vol.Spec.ClaimRef.Namespace, Name: vol.Spec.ClaimRef.Name}
		vm.mu.Lock()
		if ev == DeleteEvent {
			delete(vm.specs, name.String())
		} else {
			vm.specs[name.String()] = vol.DeepCopy()
		}
		vm.mu.Unlock()
	}
	return true
}

func (vm *VolumeManage) OnAdd(obj interface{}, isInInitialList bool) {
	if vm.handler(obj, AddEvent) {
		for _, p := range vm.proc {
			p.VolumeUpdate(&VolumeItem{
				Ev:  AddEvent,
				Vol: obj.(*corev1.PersistentVolume),
			})
		}
	}
}

func (vm *VolumeManage) OnUpdate(oldObj, newObj interface{}) {
	if vm.handler(newObj, UpdateEvent) {
		for _, p := range vm.proc {
			p.VolumeUpdate(&VolumeItem{
				Ev:  UpdateEvent,
				Vol: newObj.(*corev1.PersistentVolume),
			})
		}
	}
}

func (vm *VolumeManage) OnDelete(obj interface{}) {
	if vm.handler(obj, DeleteEvent) {
		for _, p := range vm.proc {
			p.VolumeUpdate(&VolumeItem{
				Ev:  DeleteEvent,
				Vol: obj.(*corev1.PersistentVolume),
			})
		}
	}
}
