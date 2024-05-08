package plugins

import (
	"fmt"
	"path"

	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/vmvol"
)

func init() {
	vmvol.RegistVolHandler("nfs.csi", nfsCsiHandler)
}

func nfsCsiHandler(pvs ...*vmvol.PodVol) []*vmvol.VolResult {
	var (
		ret []*vmvol.VolResult
	)
	for _, podv := range pvs {
		if podv.PvSpec == nil {
			continue
		}
		if podv.PvSpec.Spec.CSI == nil {
			continue
		}
		server := podv.PvSpec.Spec.CSI.VolumeAttributes["server"]
		share := podv.PvSpec.Spec.CSI.VolumeAttributes["share"]
		subdir := podv.PvSpec.Spec.CSI.VolumeAttributes["subdir"]
		if server == "" || share == "" || subdir == "" {
			continue
		}
		sharedir := path.Join(share, subdir)

		args := []string{"-t", "nfs"}
		args = append(args, commonOpt(&podv.PvSpec.Spec)...)
		args = append(args, fmt.Sprintf("%s:%s", server, sharedir))
		args = append(args, podv.Destination)

		data := &vmvol.VolResult{
			Destination: podv.Destination,
			Hooks: &api.Hooks{
				CreateRuntime: []*api.Hook{
					{
						Path: "mount",
						Args: args,
					},
				},
			},
		}
		ret = append(ret, data)
	}
	return ret
}
