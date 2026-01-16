package cgroup

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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

	node *util.HashMap[string, *cgroup]

	mu        sync.RWMutex
	namespace map[string]*util.HashMap[string, *cgroup]
}

func CgroupManager(no *k8s.NodeManage, ns *k8s.NsManage, po *k8s.PodManage) oci.Oci {
	m := &manager{
		po:        po,
		node:      util.New[string, *cgroup](),
		namespace: map[string]*util.HashMap[string, *cgroup]{},
	}
	no.Registe(m)
	ns.Registe(m)
	return m
}

func (m *manager) Name() string {
	return "cgroup"
}

// NOTICE. skip initContainer
func (m *manager) Process(ctx context.Context, im *oci.Item) error {
	if im == nil || im.Adjust == nil {
		return errors.Join(fmt.Errorf("process %s, but data invalid", m.Name()), &merr.InternalError{})
	}
	po, err := m.po.Pod(oci.GetPodInfo(im.Ct))
	if err != nil {
		return errors.Join(&merr.K8sError{}, err)
	}
	for _, cnt := range po.Spec.InitContainers {
		if im.Ct.Name == cnt.Name {
			klog.V(3).Infof("'%s' skip initcontainer '%s' for pod '%s/%s'", m.Name(), cnt.Name, po.GetNamespace(), po.GetName())
			return nil
		}
	}

	var (
		cpuc = &cpuCgroup{}
		memc = &memCgroup{}
		ct   = im.Ct
	)

	fn := func(cp *cgroup) {
		switch cp.Meta.(type) {
		case *cpuCgroup:
			err = cgroupMerge(cp.Meta, cpuc, true)
		case *memCgroup:
			err = cgroupMerge(cp.Meta, memc, true)
		default:
			klog.Warningf("not support cgroup type: %v", reflect.TypeOf(cp.Meta))
		}
		if err != nil {
			klog.Errorf("cgroup merge failed: %v", err)
		}
	}
	// namespace
	m.mu.RLock()
	nscg := m.namespace[oci.GetNamespace(ct)]
	m.mu.RUnlock()
	if nscg != nil {
		nscg.Iter(func(_ string, c *cgroup) bool {
			fn(c)
			return true
		})
	}

	// node
	m.node.Iter(func(_ string, c *cgroup) bool {
		fn(c)
		return true
	})

	// pod
	cg := cgroupParse(po, ct.Name)
	if cg != nil {
		fn(cg)
	}
	if cpuc.Adjust(im.Adjust) {
		klog.Infof("update pod '%s/%s', cgoup: %s", po.GetNamespace(), po.GetName(), cpuc)
	}
	if memc.Adjust(im.Adjust) {
		klog.Infof("update pod '%s/%s', cgoup: %s", po.GetNamespace(), po.GetName(), memc)
	}
	return nil
}

func (m *manager) Start(ctx context.Context, pod *api.PodSandbox, container *api.Container) error {
	return nil
}

// only support [kind].[suffix]
func (m *manager) NodeUpdate(ni *k8s.NodeItem) {
	switch ni.Ev {
	case k8s.AddEvent, k8s.UpdateEvent:
	default:
		return
	}
	node := ni.No

	for k, v := range node.Annotations {
		prefix, ok := util.TrimSuffix(k, CgroupSuffix)
		if !ok {
			continue
		}
		cg := cgroupfromStr(prefix, v)
		if cg != nil {
			m.node.Put(prefix, cg)
			klog.V(3).Infof("node '%s', update cgroup: %s", node.Name, cg)
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
		prefix, ok := util.TrimSuffix(k, CgroupSuffix)
		if !ok {
			continue
		}
		cg := cgroupfromStr(prefix, v)
		if cg != nil {
			m.mu.Lock()
			v, ok := m.namespace[ns.GetName()]
			if !ok {
				m.namespace[ns.GetName()] = util.New[string, *cgroup]()
				v = m.namespace[ns.GetName()]
			}
			v.Put(prefix, cg)
			m.mu.Unlock()
			klog.V(3).Infof("namespace '%s', update cgroup: %s", ns.Name, cg)
		}
	}
}
