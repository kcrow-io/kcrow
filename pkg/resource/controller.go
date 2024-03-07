package resource

import (
	"context"

	"github.com/containerd/nri/pkg/api"
	"github.com/yylt/kcrow/pkg/util"
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

	// po annotation by kubernetes client
	usecli bool
	cli    client.Client

	nsctl *NamespaceRsc
	noctl *NodeRsc
}

func New(ctx context.Context, nsctl *NamespaceRsc, noctl *NodeRsc, cli client.Client) *ResManage {
	ctr := &ResManage{
		ctx: ctx,
		cli: cli,

		nsctl: nsctl,
		noctl: noctl,
	}
	return ctr
}

// Priority order is pod > node > namespace
func (c *ResManage) CgroupInfo(sb *api.PodSandbox, ct *api.Container, poannotation map[string]string) *api.LinuxResources {
	if sb == nil || ct == nil {
		return nil
	}

	var (
		err  error
		cpuc = &api.LinuxCPU{}
		memc = &api.LinuxMemory{}
	)

	if ct.Linux != nil && ct.Linux.Resources != nil {
		if ct.Linux.Resources.Cpu != nil {
			cpuc = ct.Linux.Resources.Cpu
		}
		if ct.Linux.Resources.Memory != nil {
			memc = ct.Linux.Resources.Memory
		}
	}

	fn := func(cp *Cgroup) {
		switch cp.Meta.(type) {
		case *api.LinuxCPU:
			err = CgroupMerge(cp.Meta, cpuc, true)
		case *api.LinuxMemory:
			err = CgroupMerge(cp.Meta, memc, true)
		}
		if err != nil {
			klog.Errorf("cgroup merge failed: %v", err)
		}
	}
	c.nsctl.IterCgroup(sb.Namespace, func(cp *Cgroup) bool {
		fn(cp)
		return true
	})
	c.noctl.IterCgroup(func(cp *Cgroup) bool {
		fn(cp)
		return true
	})
	// highest prio
	for k, v := range poannotation {
		cp := CgroupParse(k, v)
		if cp == nil {
			continue
		}
		fn(cp)
	}

	return &api.LinuxResources{
		Cpu:    cpuc,
		Memory: memc,
	}
}

func (c *ResManage) RlimitInfo(sb *api.PodSandbox, ct *api.Container, poannotation map[string]string) []*api.POSIXRlimit {
	if sb == nil || ct == nil {
		return nil
	}
	var (
		rlimit = map[string]*Rlimit{}
	)

	fn := func(rl *Rlimit) {
		v, ok := rlimit[rl.Type]
		if !ok {
			rlimit[rl.Type] = rl
			return
		}
		//not override
		rl.Merge(v, false)
	}

	// highest prio
	for k, v := range poannotation {
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
		rs[i] = v.To()
		i++
	}
	return rs
}

func (c *ResManage) CreateContainer(ctx context.Context, p *api.PodSandbox, ct *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	var (
		annotations map[string]string
		po          = &corev1.Pod{}
	)
	nsname := getPodInfo(ct)
	if nsname.Name == "" {
		return nil, nil, nil
	}
	// NOTICE: containerd should not erase kcrow.io annotation
	// reference: https://github.com/containerd/containerd/blob/main/docs/cri/config.md#full-configuration container_annotations
	if c.usecli {
		err := c.cli.Get(c.ctx, nsname, po)
		if err != nil {
			klog.Warningf("pod %s get failed: %v", nsname, err)
			klog.Warning("skip pod annotations")
		} else {
			annotations = po.Annotations
		}
	} else {
		annotations = ct.Annotations
	}
	klog.V(2).Infof("process pod %v, annotations:%v", nsname, annotations)

	lres := c.CgroupInfo(p, ct, annotations)

	prlim := c.RlimitInfo(p, ct, annotations)
	adjust := &api.ContainerAdjustment{
		Linux: &api.LinuxContainerAdjustment{
			Resources: ct.Linux.Resources,
		},
	}
	if lres != nil {
		klog.V(2).Infof("update cpu resource %v for pod %v", lres.Cpu, nsname)
		klog.V(2).Infof("update mem resource %v for pod %v", lres.Memory, nsname)
		adjust.Linux = &api.LinuxContainerAdjustment{
			Resources: lres,
		}
	}
	if prlim != nil {
		for _, v := range prlim {
			klog.V(2).Infof("rlimit %v for pod %v", v, nsname)
		}
		adjust.Rlimits = prlim
	}
	if klog.V(2).Enabled() {
		util.Dump(ct.Name, "ContainerAdjustment", adjust)
	}
	return adjust, nil, nil
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
