package cgroup

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// ref: https://github.com/opencontainers/runtime-spec/blob/main/config-linux.md#control-groups
const (
	cgroupSuffix = ".cgroup.kcrow.io"
)

var (
	cgnames = map[string]reflect.Type{
		"cpu":    reflect.TypeOf(api.LinuxCPU{}),
		"memory": reflect.TypeOf(api.LinuxMemory{}),
	}
)

type cgroup struct {
	Type string
	Meta any
}

// reture type: *LinuxMemory, *LinuxCPU or nil.
func (c *cgroup) To() any {
	typev, ok := cgnames[c.Type]
	if !ok {
		return nil
	}
	if reflect.TypeOf(c.Meta) != typev {
		return nil
	}
	return c.Meta
}

func (c *cgroup) String() string {
	if c == nil {
		return ""
	}
	switch v := c.Meta.(type) {
	case *api.LinuxCPU:
		return v.String()
	case *api.LinuxMemory:
		return v.String()
	default:
		return ""
	}
}

func cgroupParse(po *corev1.Pod, cntname string) *cgroup {
	var (
		prefix, value string
		ok            bool
	)
	if po == nil || po.Annotations == nil {
		return nil
	}
	for k, v := range po.Annotations {
		prefix, ok = util.TrimSuffix(k, cgroupSuffix)
		if ok {
			value = v
			break
		}
	}
	kind, ok := k8s.TryParseContainer(po, cntname, prefix)
	if !ok {
		klog.V(2).Infof("skip container '%s' cgroup parse, not match", cntname)
		return nil
	}
	return cgroupfromStr(kind, value)
}

func cgroupfromStr(kind, value string) *cgroup {
	typev, ok := cgnames[kind]
	if !ok {
		klog.Errorf("not support cgroup kind: %v", kind)
		return nil
	}
	ptrvalue := reflect.New(typev).Interface()
	err := json.Unmarshal([]byte(value), ptrvalue)
	if err != nil {
		klog.Errorf("parse cgroup faild: %v", err)
		return nil
	}

	return &cgroup{
		Type: kind,
		Meta: ptrvalue,
	}
}

func cgroupMerge(src, dst any, override bool) error {
	srct := reflect.TypeOf(src)
	if srct != reflect.TypeOf(dst) && src != nil {
		return fmt.Errorf("type is not equal or is null")
	}
	switch src.(type) {
	case *api.LinuxCPU:
		cpuMerge(src.(*api.LinuxCPU), dst.(*api.LinuxCPU), override)
	case *api.LinuxMemory:
		memoryMerge(src.(*api.LinuxMemory), dst.(*api.LinuxMemory), override)
	default:
		return fmt.Errorf("not support cgroup type %v", srct)
	}
	return nil
}

// only merge cpuset, memset.
func cpuMerge(src, dst *api.LinuxCPU, override bool) {
	if src == nil || dst == nil {
		return
	}
	klog.V(2).Infof("cpuMerge src %v, dst %v, over %v", src, dst, override)
	if src.Cpus != "" {
		if dst.Cpus == "" || override {
			dst.Cpus = src.Cpus
		}
	}
	if src.Mems != "" {
		if dst.Mems == "" || override {
			dst.Mems = src.Mems
		}
	}
	if src.Quota != nil {
		if dst.Quota == nil || override {
			dst.Quota = &api.OptionalInt64{
				Value: src.Quota.Value,
			}
		}
	}
}

func memoryMerge(src, dst *api.LinuxMemory, override bool) {
	if src == nil || dst == nil {
		return
	}
	klog.V(2).Infof("memoryMerge src %v, dst %v, over %v", src, dst, override)
	if src.Reservation != nil {
		if dst.Reservation == nil || override {
			dst.Reservation = &api.OptionalInt64{
				Value: src.Reservation.Value,
			}
		}
	}
}
