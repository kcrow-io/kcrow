package vmvol

import (
	"github.com/containerd/nri/pkg/api"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type PodVol struct {
	VolumeMount corev1.VolumeMount
	Container   *api.Container
	PvSpec      *corev1.PersistentVolume
	Destination string
}

type VolResult struct {
	Destination string
	Hooks       *api.Hooks
	Device      *api.LinuxDevice
}

// NOTICE VolResult.Destination must not null
// Manager will remove mount for VolResult.Destination and append Hook, Device
type VolumeHandler func(pvs ...*PodVol) []*VolResult

var (
	volHandlers = map[string]VolumeHandler{}
)

func RegistVolHandler(name string, fn VolumeHandler) {
	klog.V(3).Infof("regist volume handler %s", name)
	volHandlers[name] = fn
}
