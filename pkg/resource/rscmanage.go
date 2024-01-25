package resource

import (
	"context"

	"github.com/containerd/nri/pkg/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (c *ResManage) CgroupInfo(sb *api.PodSandbox, ct *api.Container) *api.LinuxResources {
	if sb == nil || ct == nil {
		return nil
	}
	c.nsctl.IterCgroup(sb.Namespace, func(c *Cgroup) bool {

	})
}

func (c *ResManage) RlimitInfo(sb *api.PodSandbox, ct *api.Container) []*api.POSIXRlimit {

}
