package plugins

import (
	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/vmvol"
	linux "github.com/moby/sys/mountinfo"
	"k8s.io/klog/v2"
)

func init() {
	vmvol.RegistVolHandler("nfs.common", commonHandler)
}

func commonHandler(pvs ...*vmvol.PodVol) []*vmvol.VolResult {
	var (
		ret      []*vmvol.VolResult
		nfsFound bool
	)

	nfsFound = false
	nfsMountInfo, err := linux.GetMounts(linux.FSTypeFilter("nfs", "nfs4"))
	if err != nil {
		klog.V(3).ErrorS(err, "get mountinfo err")
		return nil
	}

	for _, podv := range pvs {
		args := []string{"mount", "-t", "nfs"}
		args = append(args, commonOpt(&podv.PvSpec.Spec)...)
		for _, m := range podv.Container.Mounts {
			if m.Destination == podv.VolumeMount.MountPath {
				if podv.VolumeMount.ReadOnly == true {
					args = append(args, "-o", "ro")
				}

				for _, info := range nfsMountInfo {
					if info.Mountpoint == m.Source {
						args = append(args, info.Source)
						nfsFound = true
						break
					}
				}

				if nfsFound {
					args = append(args, podv.Destination)
					data := &vmvol.VolResult{
						Destination: m.Destination,
						Hooks: &api.Hooks{
							CreateContainer: []*api.Hook{
								{
									Path: "/usr/bin/mkdir",
									Args: []string{podv.Destination},
								},
								{
									Path: "/usr/bin/mount",
									Args: args,
								},
							},
						},
					}
					ret = append(ret, data)
					break
				}
			}
		}

	}
	return ret
}
