package cgroup

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/oci"
	"k8s.io/klog/v2"
)

type manager struct {
	po *k8s.PodManage

	node *cgroup

	mu        sync.RWMutex
	namespace map[string]*cgroup
}

func CgroupManager(no *k8s.NodeManage, ns *k8s.NsManage, po *k8s.PodManage) oci.Oci {
	m := &manager{
		po:        po,
		namespace: map[string]*cgroup{},
	}
	no.Registe(m)
	ns.Registe(m)
	return m
}

func (m *manager) Name() string {
	return "cgroup"
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
		cpuc         = &api.LinuxCPU{}
		memc         = &api.LinuxMemory{}
		ct           = im.Ct
		poannotation = po.Annotations
	)

	if ct.Linux != nil && ct.Linux.Resources != nil {
		if ct.Linux.Resources.Cpu != nil {
			cpuc = ct.Linux.Resources.Cpu
		}
		if ct.Linux.Resources.Memory != nil {
			memc = ct.Linux.Resources.Memory
		}
	}

	fn := func(cp *cgroup) {
		switch cp.Meta.(type) {
		case *api.LinuxCPU:
			err = cgroupMerge(cp.Meta, cpuc, true)
		case *api.LinuxMemory:
			err = cgroupMerge(cp.Meta, memc, true)
		default:
			klog.Warningf("not support cgroup meta %v", reflect.TypeOf(cp.Meta))
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
		fn(nscg)
	}

	// node
	if m.node != nil {
		fn(m.node)
	}

	// pod
	for k, v := range poannotation {
		cp := cgroupParse(k, v)
		if cp == nil {
			continue
		}
		fn(cp)
	}
	ct.Linux.Resources.Cpu = cpuc
	ct.Linux.Resources.Memory = memc
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
		cg := cgroupParse(k, v)
		if cg != nil {
			klog.Infof("node %s, update cgroup %s", n.Name, cg)
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
		cg := cgroupParse(k, v)
		if cg != nil {
			m.mu.Lock()
			klog.Infof("namespace %s, update cgroup %s", n.Name, cg)
			m.namespace[n.Name] = cg
			m.mu.Unlock()
		}
	}
}
