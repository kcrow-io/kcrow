package ulimit

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/containerd/nri/pkg/api"
	merr "github.com/kcrow-io/kcrow/pkg/errors"
	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/oci"
	"github.com/kcrow-io/kcrow/pkg/util"
	"k8s.io/klog/v2"
)

type manager struct {
	po *k8s.PodManage

	node *util.HashMap[string, *rlimit]

	mu        sync.RWMutex
	namespace map[string]*util.HashMap[string, *rlimit]
}

func RlimitManager(no *k8s.NodeManage, ns *k8s.NsManage, po *k8s.PodManage) oci.Oci {
	m := &manager{
		po:        po,
		node:      util.New[string, *rlimit](),
		namespace: map[string]*util.HashMap[string, *rlimit]{},
	}
	no.Registe(m)
	ns.Registe(m)
	return m
}

func (m *manager) Name() string {
	return "ulimit"
}

func (m *manager) Start(ctx context.Context, pod *api.PodSandbox, container *api.Container) error {
	return nil
}

func (m *manager) Process(ctx context.Context, im *oci.Item) error {
	if im == nil || im.Adjust == nil {
		return errors.Join(fmt.Errorf("process %s, but data invalid", m.Name()), &merr.InternalError{})
	}
	po, err := m.po.Pod(oci.GetPodInfo(im.Ct))
	if err != nil {
		return errors.Join(&merr.K8sError{}, err)
	}
	klog.V(4).Infof("process '%s' on pod '%s/%s'", m.Name(), po.GetNamespace(), po.GetName())

	var (
		rlimitm = map[string]*rlimit{}
		ct      = im.Ct
	)

	fn := func(rl *rlimit) {
		v, ok := rlimitm[rl.Type]
		if !ok {
			rlimitm[rl.Type] = rl
			return
		}
		// override
		rl.Merge(v, true)
	}
	// namespace
	m.mu.RLock()
	nsrlimit := m.namespace[oci.GetNamespace(ct)]
	m.mu.RUnlock()
	if nsrlimit != nil {
		nsrlimit.Iter(func(_ string, c *rlimit) bool {
			fn(c)
			return true
		})
	}

	// node
	m.node.Iter(func(_ string, c *rlimit) bool {
		fn(c)
		return true
	})

	// pod
	cg := rlimitParse(po, ct.Name)
	if cg != nil {
		fn(cg)
	}

	for _, v := range rlimitm {
		apiv := v.To()
		if apiv != nil {
			klog.Infof("update pod '%s/%s' rlimit: %s", po.GetNamespace(), po.GetName(), v.String())
			im.Adjust.AddRlimit(apiv.Type, apiv.Hard, apiv.Soft)
		}
	}
	return nil
}

func (m *manager) NodeUpdate(ni *k8s.NodeItem) {
	switch ni.Ev {
	case k8s.AddEvent, k8s.UpdateEvent:
	default:
		return
	}
	node := ni.No

	for k, v := range node.Annotations {
		prefix, ok := util.TrimSuffix(k, rlimtSuffix)
		if !ok {
			continue
		}
		rlimit := rlimitfromStr(prefix, v)
		if rlimit != nil {
			m.node.Put(prefix, rlimit)
			klog.V(3).Infof("node %s, update rlimit %s", node.Name, rlimit)
		}
	}
}

// only support [kind].[suffix]
func (m *manager) NamespaceUpdate(ni *k8s.NsItem) {
	switch ni.Ev {
	case k8s.AddEvent, k8s.UpdateEvent:
	default:
		return
	}
	ns := ni.Ns

	for k, v := range ns.Annotations {
		prefix, ok := util.TrimSuffix(k, rlimtSuffix)
		if !ok {
			continue
		}
		limit := rlimitfromStr(prefix, v)
		if limit != nil {
			m.mu.Lock()
			v, ok := m.namespace[ns.GetName()]
			if !ok {
				m.namespace[ns.GetName()] = util.New[string, *rlimit]()
				v = m.namespace[ns.GetName()]
			}
			v.Put(prefix, limit)
			m.mu.Unlock()
			klog.V(3).Infof("namespace %s, update rlimit %s", ns.Name, limit)
		}
	}
}
