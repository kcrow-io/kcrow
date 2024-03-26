//go:build linux
// +build linux

package cgroup

import "github.com/containerd/nri/pkg/api"

// Must judge the upper level's content before  modified .
type CpuCgroup struct {
	Shares *uint64 `json:"shares,omitempty"`
	Quota  *int64  `json:"quota,omitempty"`
	Period *uint64 `json:"period,omitempty"`
	Cpus   string  `json:"cpus,omitempty"`
	Mems   string  `json:"mems,omitempty"`
}

func (cc *CpuCgroup) To() *api.LinuxCPU {
	linuxc := &api.LinuxCPU{}
	if cc.Cpus != "" {
		linuxc.Cpus = cc.Cpus
	}
	if cc.Mems != "" {
		linuxc.Mems = cc.Mems
	}
	if cc.Period != nil {
		linuxc.Period = &api.OptionalUInt64{
			Value: *cc.Period,
		}
	}
	if cc.Shares != nil {
		linuxc.Shares = &api.OptionalUInt64{
			Value: *cc.Shares,
		}
	}
	if cc.Quota != nil {
		linuxc.Quota = &api.OptionalInt64{
			Value: *cc.Quota,
		}
	}
	return linuxc
}

func (cc *CpuCgroup) MergeTo(dst *api.LinuxCPU, override bool) {
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
	if cc.Period != nil {
		if dst.Period == nil || override {
			dst.Period = &api.OptionalUInt64{
				Value: *cc.Period,
			}
		}
	}
	if cc.Shares != nil {
		if dst.Shares == nil || override {
			dst.Shares = &api.OptionalUInt64{
				Value: *cc.Shares,
			}
		}
	}
	if cc.Quota != nil {
		if dst.Quota == nil || override {
			dst.Quota = &api.OptionalInt64{
				Value: *cc.Quota,
			}
		}
	}
}

type MemCgroup struct {
	// sets hard limit of memory usage
	Limit *int64 `json:"limit,omitempty"`

	// sets soft limit of memory usage
	Reservation *int64 `json:"reservation,omitempty"`

	// enables or disables the OOM killer. default "false"
	DisableOomKiller *bool `json:"disableOomKiller,omitempty"`
}

func (mc *MemCgroup) To() *api.LinuxMemory {
	linuxm := &api.LinuxMemory{}
	if mc.Limit != nil {
		linuxm.Limit = &api.OptionalInt64{
			Value: *mc.Limit,
		}
	}
	if mc.Reservation != nil {
		linuxm.Reservation = &api.OptionalInt64{
			Value: *mc.Reservation,
		}
	}
	if mc.DisableOomKiller != nil {
		linuxm.DisableOomKiller = &api.OptionalBool{
			Value: *mc.DisableOomKiller,
		}
	}
	return linuxm
}

func (mc *MemCgroup) MergeTo(dst *api.LinuxMemory, override bool) {
	if dst == nil {
		return
	}
	if mc.Limit != nil {
		if dst.Limit == nil || override {
			dst.Limit = &api.OptionalInt64{
				Value: *mc.Limit,
			}
		}
	}
	if mc.Reservation != nil {
		if dst.Reservation == nil || override {
			dst.Reservation = &api.OptionalInt64{
				Value: *mc.Reservation,
			}
		}
	}
	if mc.DisableOomKiller != nil {
		if dst.DisableOomKiller == nil || override {
			dst.DisableOomKiller = &api.OptionalBool{
				Value: *mc.DisableOomKiller,
			}
		}
	}
}
