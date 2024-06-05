package plugins

import (
	"fmt"
	"os"

	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/vmvol"
	linux "github.com/moby/sys/mountinfo"
	"k8s.io/klog/v2"
)

const pid = 1

func init() {
	vmvol.RegistVolHandler("nfs.common", commonHandler)
}

func commonHandler(pvs ...*vmvol.PodVol) []*vmvol.VolResult {
	var (
		ret      []*vmvol.VolResult
		nfsFound bool
	)

	nfsFound = false
	nfsMountInfo, err := GetMount(pid, "nfs", "nfs4")
	if err != nil {
		klog.V(3).ErrorS(err, "get mountinfo err")
		return nil
	}

	for _, podv := range pvs {
		args := []string{"mount", "-t", "nfs"}
		args = append(args, commonOpt(&podv.PvSpec.Spec)...)
		for _, m := range podv.Container.Mounts {
			if m.Destination == podv.VolumeMount.MountPath {
				if podv.VolumeMount.ReadOnly {
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
									Args: []string{"mkdir", "-p", podv.Destination},
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

func GetMount(pid int64, fstype ...string) ([]*linux.Info, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/mountinfo", pid))
	if err != nil {
		klog.V(3).ErrorS(err, "open mountinfo err")
		return nil, err
	}
	defer f.Close()
	return linux.GetMountsFromReader(f, linux.FSTypeFilter(fstype...))
}
