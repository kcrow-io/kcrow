//go:build linux
// +build linux

package cgroup

import (
	"fmt"

	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/util"
)

// Must judge the upper level's content before  modified .
type cpuCgroup struct {
	Cpus string `json:"cpus,omitempty"`
	Mems string `json:"mems,omitempty"`
}

func (cc *cpuCgroup) MergeTo(dst *cpuCgroup, override bool) {
	if dst == nil {
		return
	}
	if cc.Cpus != "" {
		if dst.Cpus == "" || override {
			dst.Cpus = cc.Cpus
		}
	}
	if cc.Mems != "" {
		if dst.Mems == "" || override {
			dst.Mems = cc.Mems
		}
	}
}

func (cc *cpuCgroup) String() string {
	buf := util.GetBuf()
	defer util.PutBuf(buf)

	buf.WriteString("cpu{")
	if cc.Cpus != "" {
		buf.WriteString(fmt.Sprintf("cpus:%s, ", cc.Cpus))
	}
	if cc.Mems != "" {
		buf.WriteString(fmt.Sprintf("mems:%v", cc.Mems))
	}
	buf.WriteString("}")
	return buf.String()
}

func (cc *cpuCgroup) Adjust(aj *api.ContainerAdjustment) bool {
	if cc.Cpus == "" && cc.Mems == "" {
		return false
	}
	if cc.Cpus != "" {
		aj.SetLinuxCPUSetCPUs(cc.Cpus)
	}
	if cc.Mems != "" {
		aj.SetLinuxCPUSetMems(cc.Mems)
	}
	return true
}

type memCgroup struct {
	// sets soft limit of memory usage
	Reservation *int64 `json:"reservation,omitempty"`

	// enables or disables the OOM killer. default "false"
	DisableOomKiller *bool `json:"disableOomKiller,omitempty"`
}

func (mc *memCgroup) String() string {
	buf := util.GetBuf()
	defer util.PutBuf(buf)

	buf.WriteString("memory{")
	if mc.Reservation != nil {
		buf.WriteString(fmt.Sprintf("reservation:%d, ", mc.Reservation))
	}
	if mc.DisableOomKiller != nil {
		buf.WriteString(fmt.Sprintf("disableOomKiller:%v", *mc.DisableOomKiller))
	}
	buf.WriteString("}")
	return buf.String()
}

func (mc *memCgroup) Adjust(aj *api.ContainerAdjustment) bool {
	if mc.DisableOomKiller == nil && mc.Reservation == nil {
		return false
	}
	if mc.DisableOomKiller != nil && *mc.DisableOomKiller {
		aj.SetLinuxMemoryDisableOomKiller()
	}
	if mc.Reservation != nil {
		aj.SetLinuxMemoryReservation(*mc.Reservation)
	}
	return true
}

func (mc *memCgroup) MergeTo(dst *memCgroup, override bool) {
	if dst == nil {
		return
	}
	if mc.Reservation != nil {
		if dst.Reservation == nil || override {
			dst.Reservation = mc.Reservation
		}
	}
	if mc.DisableOomKiller != nil {
		if dst.DisableOomKiller == nil || override {
			dst.DisableOomKiller = mc.DisableOomKiller
		}
	}
}
