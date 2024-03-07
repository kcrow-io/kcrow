package resource

import (
	"context"
	"os"
	"sync"

	corev1 "k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

var (
	nodeName = os.Getenv("NODE_NAME")
)

type NodeRsc struct {
	ctx context.Context

	reader cache.Cache

	cgmu sync.RWMutex
	cg   map[string]*Cgroup

	rlmu   sync.RWMutex
	rlimit map[string]*Rlimit

	syncedFn func() bool
}

// only record current node
func NewNodeControl(ctx context.Context, reader cache.Cache) *NodeRsc {
	no := &NodeRsc{
		reader: reader,
		cg:     map[string]*Cgroup{},
		rlimit: map[string]*Rlimit{},
	}
	err := no.probe()
	if err != nil {
		panic(err)
	}
	return no
}

// priority
func (no *NodeRsc) probe() error {
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

func (no *NodeRsc) add(n *corev1.Node) {
	for k, v := range n.Annotations {
		cg := CgroupParse(k, v)
		if cg != nil {
			klog.Infof("node %s, cgroup %v", n.Name, cg)
			no.cgmu.Lock()
			no.cg[cg.Type] = cg
			no.cgmu.Unlock()
		}
		rl := RlimitParse(k, v)
		if rl != nil {
			klog.Infof("node %s, rlimit %v", n.Name, rl)
			no.rlmu.Lock()
			no.rlimit[rl.Type] = rl
			no.rlmu.Unlock()
		}
	}
}

func (no *NodeRsc) OnAdd(obj interface{}, isInInitialList bool) {
	no.add(obj.(*corev1.Node))
}

func (no *NodeRsc) OnUpdate(oldObj, newObj interface{}) {
	no.add(newObj.(*corev1.Node))
}

// do nothing
func (no *NodeRsc) OnDelete(obj interface{}) {

}

func (no *NodeRsc) IterCgroup(fn func(*Cgroup) bool) {
	if !no.syncedFn() {
		return
	}

	no.cgmu.RLock()
	defer no.cgmu.RUnlock()
	for _, v := range no.cg {
		ok := fn(v)
		if !ok {
			return
		}
	}
}

func (no *NodeRsc) IterRlimit(fn func(*Rlimit) bool) {
	if !no.syncedFn() {
		return
	}

	no.rlmu.RLock()
	defer no.rlmu.RUnlock()
	for _, v := range no.rlimit {
		ok := fn(v)
		if !ok {
			return
		}
	}
}
