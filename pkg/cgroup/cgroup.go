package cgroup

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// ref: https://github.com/opencontainers/runtime-spec/blob/main/config-linux.md#control-groups
const (
	CgroupSuffix = ".cgroup.kcrow.io"
)

var (
	cgnames = map[string]reflect.Type{
		"cpu": reflect.TypeOf(cpuCgroup{}),
		"mem": reflect.TypeOf(memCgroup{}),
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
	case *cpuCgroup:
		return v.String()
	case *memCgroup:
		return v.String()
	default:
		return ""
	}
}

func cgroupParse(po *corev1.Pod, cntname string) *cgroup {
	var (
		prefix, value string
		found         bool
	)
	if po == nil || po.Annotations == nil {
		return nil
	}
	for k, v := range po.Annotations {
		prefix, found = util.TrimSuffix(k, CgroupSuffix)
		if found {
			value = v
			break
		}
	}
	if !found {
		return nil
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
	case *cpuCgroup:
		src.(*cpuCgroup).MergeTo(dst.(*cpuCgroup), override)
	case *memCgroup:
		src.(*memCgroup).MergeTo(dst.(*memCgroup), override)
	default:
		return fmt.Errorf("not support cgroup type: %v", srct)
	}
	return nil
}
