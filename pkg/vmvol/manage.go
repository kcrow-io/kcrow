package vmvol

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/containerd/nri/pkg/api"
	merr "github.com/kcrow-io/kcrow/pkg/errors"
	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/oci"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

const (
	name = "vm-volume"
)

type manager struct {
	ctx context.Context

	pm   *k8s.PodManage
	volm *k8s.VolumeManage
	rcm  *k8s.RuntimeManage
}

func New(ctx context.Context, volm *k8s.VolumeManage, rcm *k8s.RuntimeManage, pm *k8s.PodManage) oci.Oci {
	m := &manager{
		ctx:  ctx,
		pm:   pm,
		volm: volm,
		rcm:  rcm,
	}
	return m
}

func (m *manager) Name() string {
	return name
}

func (m *manager) Process(ctx context.Context, im *oci.Item) error {
	if im == nil || im.Adjust == nil {
		return errors.Join(fmt.Errorf("process %s, but data invalid", m.Name()), &merr.InternalError{})
	}
	po, err := m.pm.Pod(oci.GetPodInfo(im.Ct))
	if err != nil {
		return errors.Join(&merr.K8sError{}, err)
	}
	// generally, default runtime is runc
	if po.Spec.RuntimeClassName == nil {
		return nil
	}
	// when runtimeClassName is kata, we need to handle the volume
	if po.Spec.RuntimeClassName != nil && !m.rcm.IsKata(*po.Spec.RuntimeClassName) {
		return nil
	}

	// record PersistentVolumeClaim
	var (
		vols    = map[string]*corev1.PersistentVolume{}
		podvols []*PodVol
	)
	for _, vol := range po.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}

		pvc := types.NamespacedName{
			Namespace: po.GetNamespace(),
			Name:      vol.PersistentVolumeClaim.ClaimName,
		}

		vols[vol.Name] = m.volm.GetVolumeSpec(pvc)
	}

	for _, cnt := range po.Spec.Containers {
		if cnt.Name != im.Ct.Name {
			continue
		}

		for _, volmnt := range cnt.VolumeMounts {
			v, ok := vols[volmnt.Name]
			if !ok {
				continue
			}
			// for kata runtime, the destination has prefix
			// /run/kata-containers/$cid/rootfs
			podvols = append(podvols, &PodVol{
				VolumeMount: volmnt,
				Container:   im.Ct,
				PvSpec:      v,
				Destination: path.Join(kataPrefixPath(im.Ct), volmnt.MountPath),
			})

		}
	}
	if len(podvols) != 0 {
		klog.V(4).Infof("find vm runtime persistent volume info: %v", podvols)
		for name, hand := range volHandlers {
			klog.V(3).Infof("start volumehandle '%s' on container %s", name, im.Ct.Name)
			results := hand(podvols...)
			for _, res := range results {
				klog.V(3).Infof("container %s, volume transfer info: %v", im.Ct.Name, res)
				if res.Hooks != nil {
					im.Adjust.AddHooks(res.Hooks)
				}
				if res.Device != nil {
					im.Adjust.AddDevice(res.Device)
				}
				im.Adjust.RemoveMount(res.Destination)
			}
		}
	}
	return nil
}

func (m *manager) Start(ctx context.Context, pod *api.PodSandbox, container *api.Container) error {
	return nil
}

func kataPrefixPath(cnt *api.Container) string {
	return fmt.Sprintf("/run/kata-containers/%s/rootfs", cnt.Id)
}
