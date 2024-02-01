package resource

import (
	"context"

	"github.com/containerd/nri/pkg/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PodName      = "io.kubernetes.pod.name"
	PodNamespace = "io.kubernetes.pod.namespace"
)

type ResManage struct {
	ctx context.Context

	client.Client

	nsctl NamespaceRsc
	noctl NodeRsc
}

func New(nsctl *NamespaceRsc, noctl *NodeRsc) *ResManage {
	ctr := &ResManage{}
	return ctr
}

// Priority order is pod > node > namespace
func (c *ResManage) CgroupInfo(sb *api.PodSandbox, ct *api.Container) *api.LinuxResources {
	if sb == nil || ct == nil {
		return nil
	}
	// api.Container annotations is not equal pod
	nsname := getPodInfo(ct)
	if nsname.Name == "" {
		return nil
	}
	var (
		po = &corev1.Pod{}

		cpuc = &api.LinuxCPU{}
		memc = &api.LinuxMemory{}
	)
	err := c.Client.Get(c.ctx, nsname, po)
	if err != nil {
		klog.Errorf("pod %s get failed: %v", nsname, err)
		return nil
	}
	fn := func(cp *Cgroup) {
		//not override
		switch cp.Type {
		case CGROUP_CPU:
			v, ok := cp.Meta.(*CpuCgroup)
			if ok {
				v.MergeTo(cpuc, false)
			}
		case CGROUP_MEM:
			v, ok := cp.Meta.(*MemCgroup)
			if ok {
				v.MergeTo(memc, false)
			}
		}
	}

	for k, v := range po.Annotations {
		cp := CgroupParse(k, v)
		if cp == nil {
			continue
		}
		fn(cp)
	}
	c.noctl.IterCgroup(func(cp *Cgroup) bool {
		fn(cp)
		return true
	})
	c.nsctl.IterCgroup(sb.Namespace, func(cp *Cgroup) bool {
		fn(cp)
		return true
	})
	return &api.LinuxResources{
		Cpu:    cpuc,
		Memory: memc,
	}
}

func (c *ResManage) RlimitInfo(sb *api.PodSandbox, ct *api.Container) []*api.POSIXRlimit {
	if sb == nil || ct == nil {
		return nil
	}
	// api.Container annotations is not equal pod
	nsname := getPodInfo(ct)
	if nsname.Name == "" {
		return nil
	}
	var (
		po     = &corev1.Pod{}
		rlimit = map[RlimitRsc]*Rlimit{}
	)
	err := c.Client.Get(c.ctx, nsname, po)
	if err != nil {
		klog.Errorf("pod %s get failed: %v", nsname, err)
		return nil
	}
	fn := func(rl *Rlimit) {
		v, ok := rlimit[rl.Type]
		if !ok {
			rlimit[rl.Type] = rl
			return
		}
		//not override
		rl.Merge(v, false)
	}
	for k, v := range po.Annotations {
		cp := RlimitParse(k, v)
		if cp == nil {
			continue
		}
		fn(cp)
	}
	c.noctl.IterRlimit(func(cp *Rlimit) bool {
		fn(cp)
		return true
	})
	c.nsctl.IterRlimit(sb.Namespace, func(cp *Rlimit) bool {
		fn(cp)
		return true
	})
	var (
		rs = make([]*api.POSIXRlimit, len(rlimit))
		i  = 0
	)
	for _, v := range rlimit {
		rs[i] = v.Resource()
		i++
	}
	return rs
}

func getPodInfo(ct *api.Container) types.NamespacedName {
	if ct == nil || ct.Labels == nil {
		return types.NamespacedName{}
	}
	return types.NamespacedName{
		Namespace: ct.Labels[PodNamespace],
		Name:      ct.Labels[PodName],
	}
}
