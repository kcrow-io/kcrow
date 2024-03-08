package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/containerd/nri/pkg/api"
	"k8s.io/klog/v2"
)

// ref: https://github.com/opencontainers/runtime-spec/blob/main/config-linux.md#control-groups
const (
	CgroupSuffix = ".cgroup.kcrow.io"
)

var (
	cgnames = map[string]reflect.Type{
		"cpu":    reflect.TypeOf(api.LinuxCPU{}),
		"memory": reflect.TypeOf(api.LinuxMemory{}),
	}
)

type Cgroup struct {
	Type string
	Meta any
}

// reture type: *LinuxMemory, *LinuxCPU or nil.
func (c *Cgroup) To() any {
	typev, ok := cgnames[c.Type]
	if !ok {
		return nil
	}
	if reflect.TypeOf(c.Meta) != typev {
		return nil
	}
	return c.Meta
}

func (c *Cgroup) String() string {
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

func CgroupParse(key, value string) *Cgroup {
	var (
		idx int
	)
	if idx = strings.Index(key, CgroupSuffix); idx < 0 {
		return nil
	}

	// TODO support select container.
	kind := strings.ToLower(key[:idx])

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

	return &Cgroup{
		Type: kind,
		Meta: ptrvalue,
	}
}

func CgroupMerge(src, dst any, override bool) error {
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
	if src.Period != nil {
		if dst.Period == nil || override {
			dst.Period = &api.OptionalUInt64{
				Value: src.Period.Value,
			}
		}
	}
	if src.Shares != nil {
		if dst.Shares == nil || override {
			dst.Shares = &api.OptionalUInt64{
				Value: src.Shares.Value,
			}
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
	if src.DisableOomKiller != nil {
		if dst.DisableOomKiller == nil || override {
			dst.DisableOomKiller = &api.OptionalBool{
				Value: src.DisableOomKiller.Value,
			}
		}
	}
	if src.Limit != nil {
		if dst.Limit == nil || override {
			dst.Limit = &api.OptionalInt64{
				Value: src.Limit.Value,
			}
		}
	}
}
