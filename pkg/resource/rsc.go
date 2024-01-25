package resource

import (
	"encoding/json"
	"strings"

	"github.com/containerd/nri/pkg/api"
)

const (
	RlimtPrefix = "rlimit.kcrow.io/"

	CgroupPrefix = "cgroup.kcrow.io/"
)

// ref: https://github.com/opencontainers/runtime-spec/blob/main/config.md#posix-process
type RlimitRsc string

type Rlimit struct {
	Type RlimitRsc `json:"-"`
	Hard *uint64   `json:"hard,omitempty"`
	Soft *uint64   `json:"soft,omitempty"`
}

const (
	RLIMIT_MEMLOCK RlimitRsc = "memlock"
	RLIMIT_NOFILE  RlimitRsc = "nofile"
	RLIMIT_NPROC   RlimitRsc = "nproc"
	RLIMIT_CORE    RlimitRsc = "core"
)

// ref: https://github.com/opencontainers/runtime-spec/blob/main/config-linux.md#control-groups
type CgroupRsc string

type Cgroup struct {
	Type CgroupRsc
	Meta any
}

func (r *Rlimit) Merge(dst *Rlimit, override bool) {
	if dst == nil {
		return
	}
	if r.Type != dst.Type {
		return
	}

	if r.Hard != nil {
		if dst.Hard == nil || override {
			dst.Hard = r.Hard
		}
	}
	if r.Soft != nil {
		if dst.Soft == nil || override {
			dst.Soft = r.Soft
		}
	}
}

// when one of hard/soft is null
// will use non-null to override.
func (r *Rlimit) Adjust() {
	if r.Hard != nil && r.Soft != nil {
		return
	}
	if r.Hard == nil && r.Soft == nil {
		return
	}
	if r.Hard != nil {
		r.Soft = r.Hard
	} else {
		r.Hard = r.Soft
	}
}

func (r *Rlimit) Resource() *api.POSIXRlimit {
	var (
		ret = &api.POSIXRlimit{}
	)
	switch r.Type {
	case (RLIMIT_MEMLOCK):
		ret.Type = "RLIMIT_MEMLOCK"
	case (RLIMIT_CORE):
		ret.Type = "RLIMIT_CORE"
	case (RLIMIT_NOFILE):
		ret.Type = "RLIMIT_NOFILE"
	case (RLIMIT_NPROC):
		ret.Type = "RLIMIT_NPROC"
	default:
		return nil
	}
	if r.Soft == nil && r.Hard == nil {
		return nil
	}

	if r.Soft != nil {
		if r.Hard != nil {
			if *r.Hard < *r.Soft {
				return nil
			}
		} else {
			r.Hard = r.Soft
		}
	} else {
		r.Soft = r.Hard
	}

	ret.Hard = *r.Hard
	ret.Soft = *r.Soft
	return ret
}

func (c *Cgroup) Resource() any {
	switch c.Type {
	case (CGROUP_CPU):

	case (CGROUP_MEM):

	default:
		return nil
	}
}

func RlimitParse(key, value string) *Rlimit {

	if !strings.HasPrefix(key, RlimtPrefix) {
		return nil
	}

	kind := key[len(RlimtPrefix):]
	ret := &Rlimit{}

	switch strings.ToLower(kind) {
	case string(RLIMIT_MEMLOCK):
		ret.Type = RLIMIT_MEMLOCK
	case string(RLIMIT_CORE):
		ret.Type = RLIMIT_CORE
	case string(RLIMIT_NOFILE):
		ret.Type = RLIMIT_NOFILE
	case string(RLIMIT_NPROC):
		ret.Type = RLIMIT_NPROC
	default:
		return nil
	}
	err := json.Unmarshal([]byte(value), ret)
	if err != nil {
		return nil
	}
	// must set one.
	if ret.Hard == nil && ret.Soft == nil {
		return nil
	}
	return ret
}

func CgroupParse(key, value string) *Cgroup {

	if !strings.HasPrefix(key, CgroupPrefix) {
		return nil
	}

	kind := key[len(CgroupPrefix):]
	ret := &Cgroup{}

	switch strings.ToLower(kind) {
	case string(CGROUP_CPU):
		ret.Type = CGROUP_CPU
		meta := &CpuCgroup{}
		err := json.Unmarshal([]byte(value), meta)
		if err != nil {
			return nil
		}
		ret.Meta = meta
	case string(CGROUP_MEM):
		ret.Type = CGROUP_MEM
		meta := &MemCgroup{}
		err := json.Unmarshal([]byte(value), meta)
		if err != nil {
			return nil
		}
		ret.Meta = meta
	}

	return ret
}
