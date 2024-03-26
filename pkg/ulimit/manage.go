package ulimit

import (
	"context"
	"fmt"
	"sync"

	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/oci"
	"k8s.io/klog/v2"
)

type manager struct {
	po *k8s.PodManage

	node *rlimit

	mu        sync.RWMutex
	namespace map[string]*rlimit
}

func RlimitManager(no *k8s.NodeManage, ns *k8s.NsManage, po *k8s.PodManage) oci.Oci {
	m := &manager{
		po:        po,
		namespace: map[string]*rlimit{},
	}
	no.Registe(m)
	ns.Registe(m)
	return m
}

func (m *manager) Name() string {
	return "ulimit"
}

func (m *manager) Process(ctx context.Context, im *oci.Item) error {
	if im == nil || im.Adjust == nil {
		klog.Warningf("cgroup could not process nil data")
		return fmt.Errorf("oci data must be set")
	}
	po, err := m.po.Pod(oci.GetPodInfo(im.Ct))
	if err != nil {
		return err
	}

	var (
		rlimitm      = map[string]*rlimit{}
		poannotation = po.Annotations
		ct           = im.Ct
	)

	fn := func(rl *rlimit) {
		v, ok := rlimitm[rl.Type]
		if !ok {
			rlimitm[rl.Type] = rl
			return
		}
		//not override
		rl.Merge(v, false)
	}

	// namespace
	m.mu.RLock()
	nscg := m.namespace[oci.GetNamespace(ct)]
	m.mu.RUnlock()
	fn(nscg)

	// node
	fn(m.node)

	// pod
	for k, v := range poannotation {
		cp := rlimitParse(k, v)
		if cp == nil {
			continue
		}
		fn(cp)
	}
	for _, v := range rlimitm {
		im.Adjust.Rlimits = append(im.Adjust.Rlimits, v.To())
	}
	return nil
}

func (m *manager) NodeUpdate(ni *k8s.NodeItem) {
	switch ni.Ev {
	case k8s.AddEvent, k8s.UpdateEvent:
	default:
		return
	}
	n := ni.No
	for k, v := range n.Annotations {
		cg := rlimitParse(k, v)
		if cg != nil {
			klog.Infof("node %s, update rlimit %s", n.Name, cg)
			m.node = cg
		}
	}
}

func (m *manager) NamespaceUpdate(ni *k8s.NsItem) {
	switch ni.Ev {
	case k8s.AddEvent, k8s.UpdateEvent:
	default:
		return
	}
	n := ni.Ns
	for k, v := range n.Annotations {
		cg := rlimitParse(k, v)
		if cg != nil {
			m.mu.Lock()
			klog.Infof("namespace %s, update rlimit %s", n.Name, cg)
			m.namespace[n.Name] = cg
			m.mu.Unlock()
		}
	}
}
