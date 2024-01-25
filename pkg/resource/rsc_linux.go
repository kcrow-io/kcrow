//go:build linux
// +build linux

package resource

const (
	CGROUP_CPU CgroupRsc = "cpu"

	CGROUP_MEM CgroupRsc = "mem"
)

// Must calculate the upper level's content before whether it can be modified.
type CpuCgroup struct {
	Shares *uint64 `json:"shares,omitempty"`
	Quota  *int64  `json:"quota,omitempty"`
	Period *uint64 `json:"period,omitempty"`
	Cpus   string  `json:"cpus,omitempty"`
	Mems   string  `json:"mems,omitempty"`
}

type MemCgroup struct {
	Limit            *int64 `json:"limit,omitempty"`
	Reservation      *int64 `json:"reservation,omitempty"`
	Swap             *int64 `json:"swap,omitempty"`
	DisableOomKiller *bool  `json:"disableOomKiller,omitempty"`
}
