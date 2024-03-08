package resource

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/containerd/nri/pkg/api"
	"github.com/yylt/kcrow/pkg/util"
	"k8s.io/klog/v2"
)

// ref: https://github.com/opencontainers/runtime-spec/blob/main/config.md#posix-process

type Rlimit struct {
	Type string  `json:"-"`
	Hard *uint64 `json:"hard,omitempty"`
	Soft *uint64 `json:"soft,omitempty"`
}

const (
	RlimtSuffix = ".rlimit.kcrow.io"
)

var (
	rlimitNames = map[string]struct{}{
		"AS":         {},
		"CORE":       {},
		"CPU":        {},
		"DATA":       {},
		"FSIZE":      {},
		"LOCKS":      {},
		"MEMLOCK":    {},
		"MSGQUEUE":   {},
		"NICE":       {},
		"NOFILE":     {},
		"NPROC":      {},
		"RSS":        {},
		"RTPRIO":     {},
		"RTTIME":     {},
		"SIGPENDING": {},
		"STACK":      {},
	}
)

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
func (r *Rlimit) String() string {
	if r == nil {
		return ""
	}
	buf := util.GetBuf()
	buf.WriteString(r.Type)
	if r.Hard != nil {
		buf.WriteString(fmt.Sprintf(" %d(hard)", *r.Hard))
	}
	if r.Soft != nil {
		buf.WriteString(fmt.Sprintf(" %d(soft)", *r.Soft))
	}
	defer util.PutBuf(buf)
	return buf.String()
}

func (r *Rlimit) To() *api.POSIXRlimit {
	var (
		ret = &api.POSIXRlimit{
			Type: "RLIMIT_" + r.Type,
		}
	)
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

func RlimitParse(key, value string) *Rlimit {
	var (
		idx int
		ok  bool
	)
	if idx = strings.Index(key, RlimtSuffix); idx < 0 {
		return nil
	}
	// TODO select container.
	kind := key[:idx]
	ret := &Rlimit{}
	lk := strings.ToUpper(kind)
	_, ok = rlimitNames[lk]
	if !ok {
		klog.Errorf("not support rlimit kind: %v", kind)
		return nil
	}
	ret.Type = lk
	err := resolvRlimit(value, ret)
	if err != nil {
		klog.Errorf("parse rlimit %v faild: %v", ret.Type, err)
		return nil
	}
	// must set one.
	err = validRlimit(ret)
	if err != nil {
		klog.Errorf("rlimit %v invalid: %v", ret.Type, err)
		return nil
	}
	return ret
}

func resolvRlimit(value string, r *Rlimit) error {
	num, err := strconv.ParseUint(value, 10, 64)
	if err == nil {
		r.Hard = &num
		r.Soft = &num
		return nil
	}
	return json.Unmarshal([]byte(value), r)
}

func validRlimit(r *Rlimit) error {
	if r == nil {
		return fmt.Errorf("rlimit is null")
	}
	if r.Hard == nil && r.Soft == nil {
		return fmt.Errorf("ulimit %q must have hard limit >= soft limit", r.Type)
	}
	if r.Hard != nil && r.Soft != nil {
		if *r.Hard < *r.Soft {
			return fmt.Errorf("ulimit %q must have hard limit >= soft limit", r.Type)
		}
	}

	return nil
}
