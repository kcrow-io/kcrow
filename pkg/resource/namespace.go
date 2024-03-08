package resource

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type nsResource struct {
	cgmu sync.RWMutex
	cg   map[string]*Cgroup

	rlmu   sync.RWMutex
	rlimit map[string]*Rlimit
}

type NamespaceRsc struct {
	ctx context.Context

	reader cache.Cache

	mu        sync.RWMutex
	resources map[string]*nsResource

	HadSynced func() bool
}

func NewNsControl(ctx context.Context, reader cache.Cache) *NamespaceRsc {
	nr := &NamespaceRsc{
		reader:    reader,
		resources: map[string]*nsResource{},
	}
	err := nr.probe()
	if err != nil {
		panic(err)
	}
	return nr
}

// priority
func (nr *NamespaceRsc) probe() error {
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

func (nr *NamespaceRsc) add(no *corev1.Namespace) {
	nr.mu.Lock()
	res, ok := nr.resources[no.Name]
	if !ok {
		klog.Infof("add new namespace %s", no.Name)
		nr.resources[no.Name] = &nsResource{
			cg:     map[string]*Cgroup{},
			rlimit: map[string]*Rlimit{},
		}
		res = nr.resources[no.Name]
	}
	nr.mu.Unlock()

	for k, v := range no.Annotations {
		cg := CgroupParse(k, v)
		if cg != nil {
			klog.Infof("namespace %s, cgroup %s", no.Name, cg.Meta)
			res.cgmu.Lock()
			res.cg[cg.Type] = cg
			res.cgmu.Unlock()
		}
		rl := RlimitParse(k, v)
		if rl != nil {
			klog.Infof("namespace %s, rlimit %s", no.Name, rl)
			res.rlmu.Lock()
			res.rlimit[rl.Type] = rl
			res.rlmu.Unlock()
		}
	}
}

func (nr *NamespaceRsc) delete(no *corev1.Namespace) {
	nr.mu.Lock()
	defer nr.mu.Unlock()
	res, ok := nr.resources[no.Name]
	if !ok {
		return
	}
	res.rlmu.Lock()
	clear(res.rlimit)
	res.rlmu.Unlock()

	res.cgmu.Lock()
	clear(res.cg)
	res.cgmu.Unlock()
}

func (nr *NamespaceRsc) OnAdd(obj interface{}, isInInitialList bool) {

	nr.add(obj.(*corev1.Namespace))
}

func (nr *NamespaceRsc) OnUpdate(oldObj, newObj interface{}) {
	nr.add(newObj.(*corev1.Namespace))
}

func (nr *NamespaceRsc) OnDelete(obj interface{}) {
	nr.delete(obj.(*corev1.Namespace))
}

func (nr *NamespaceRsc) IterCgroup(ns string, fn func(*Cgroup) bool) {
	if !nr.HadSynced() {
		return
	}

	nr.mu.RLock()
	res, ok := nr.resources[ns]
	nr.mu.RUnlock()
	if !ok {
		return
	}
	res.cgmu.RLock()
	defer res.cgmu.RUnlock()
	for _, v := range res.cg {
		ok := fn(v)
		if !ok {
			return
		}
	}
}

func (nr *NamespaceRsc) IterRlimit(ns string, fn func(*Rlimit) bool) {
	if !nr.HadSynced() {
		return
	}

	nr.mu.RLock()
	res, ok := nr.resources[ns]
	nr.mu.RUnlock()
	if !ok {
		return
	}
	res.rlmu.RLock()
	defer res.rlmu.RUnlock()
	for _, v := range res.rlimit {
		ok := fn(v)
		if !ok {
			return
		}
	}
}
