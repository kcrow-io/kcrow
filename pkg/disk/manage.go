package disk

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/oci"
	"k8s.io/klog/v2"
)

const (
	DiskAnnotation      = "size.disk.kcorw.io"
	ContainerdBasePath  = "/run/containerd/io.containerd.runtime.v2.task/k8s.io/"
	ContainerdRootPath  = "/var/lib/containerd"
	SystemMountInfoFile = "/proc/1/mountinfo"
)

type manager struct {
	po        *k8s.PodManage
	mu        sync.RWMutex
	namespace map[string]string
}

func DiskManager(ns *k8s.NsManage, po *k8s.PodManage) oci.Oci {
	if !checkContainerdRootPathQuotaEnabled() {
		klog.Warning("The disk where /var/lib/containerd is located does not have prjquota enabled, skipping DiskManager init..... ")
		return nil
	}

	m := &manager{
		po:        po,
		namespace: make(map[string]string),
	}
	ns.Registe(m)
	return m
}

func (m *manager) Name() string { return "disk" }

func (m *manager) Process(ctx context.Context, im *oci.Item) error {
	return nil
}

func (m *manager) Start(ctx context.Context, pod *api.PodSandbox, container *api.Container) error {

	limitStr, ok := pod.Annotations[DiskAnnotation]
	if !ok {
		limitStr, ok = m.namespace[pod.Namespace]
		if !ok {
			return nil
		}
	}

	limitMB, err := strconv.Atoi(limitStr)
	if err != nil || limitMB <= 0 {
		klog.Errorf("Invalid quota limit for %s: %s", pod.Name, limitStr)
		return nil
	}

	rootfsPath := filepath.Join(ContainerdBasePath, container.Id, "rootfs")

	klog.V(2).Infof("Applying quota %d MB to container %s (ID: %s) at %s", limitMB, container.Name, container.Id, rootfsPath)

	runPath := filepath.Join(ContainerdBasePath, container.Id, "rootfs")
	//Obtain the snapshot ID of overlays as the ProjectID of xfs_quota
	snapshotID, foundPath, err := getOverlayPath(runPath)
	if err == nil && foundPath != "" {
		rootfsPath = foundPath
	} else {
		klog.Errorf("Could not find physical path for container %s", container.Id)
		return nil
	}

	klog.V(2).Infof("Target XFS Quota Path: %s, Quota ProjectID: %v", rootfsPath, snapshotID)

	if err := applyXFSQuota(snapshotID, rootfsPath, limitMB); err != nil {
		klog.Errorf("Failed to apply quota: %v", err)
	}
	return nil
}

func (m *manager) NamespaceUpdate(ni *k8s.NsItem) {
	switch ni.Ev {
	case k8s.AddEvent, k8s.UpdateEvent:
	default:
		return
	}
	val := ni.Ns.Annotations[DiskAnnotation]
	m.mu.Lock()
	defer m.mu.Unlock()
	if val != "" {
		m.namespace[ni.Ns.GetName()] = val
	} else {
		delete(m.namespace, ni.Ns.GetName())
	}
}
