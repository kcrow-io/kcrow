package ulimit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// ref: https://github.com/opencontainers/runtime-spec/blob/main/config.md#posix-process

// NOTICE. Soft must little than Hard.
type rlimit struct {
	Type string  `json:"-"`
	Hard *uint64 `json:"hard,omitempty"`
	Soft *uint64 `json:"soft,omitempty"`
}

const (
	rlimtSuffix = ".rlimit.kcrow.io"
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

func (r *rlimit) Merge(dst *rlimit, override bool) {
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

func (r *rlimit) String() string {
	if r == nil {
		return ""
	}
	buf := util.GetBuf()
	buf.WriteString(r.Type)
	if r.Hard != nil {
		buf.WriteString(fmt.Sprintf("-%d", *r.Hard))
	}
	if r.Soft != nil {
		buf.WriteString(fmt.Sprintf("-%d", *r.Soft))
	}
	defer util.PutBuf(buf)
	return buf.String()
}

func (r *rlimit) To() *api.POSIXRlimit {
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
				klog.Warningf("rlimit type '%s', soft '%d' will override hard '%d'", r.Type, *r.Soft, *r.Hard)
				r.Hard = r.Soft
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

func rlimitParse(po *corev1.Pod, cntname string) *rlimit {
	var (
		prefix, value string
		found         bool
	)
	if po == nil || po.Annotations == nil {
		return nil
	}
	for k, v := range po.Annotations {
		prefix, found = util.TrimSuffix(k, rlimtSuffix)
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
	return rlimitfromStr(kind, value)
}

func rlimitfromStr(kind, value string) *rlimit {
	ret := &rlimit{}
	lk := strings.ToUpper(kind)
	_, ok := rlimitNames[lk]
	if !ok {
		klog.Warningf("not support rlimit kind: %v", kind)
		return nil
	}

	ret.Type = lk
	err := resolvRlimit(value, ret)
	if err != nil {
		klog.Warningf("parse rlimit '%v' faild", value)
		return nil
	}

	err = validRlimit(ret)
	if err != nil {
		klog.Warningf("rlimit %v is invalid: %v", ret.Type, err)
		return nil
	}
	return ret
}

func resolvRlimit(value string, r *rlimit) error {
	num, err := strconv.ParseUint(value, 10, 64)
	if err == nil {
		r.Hard = &num
		r.Soft = &num
		return nil
	}
	return json.Unmarshal([]byte(value), r)
}

// hard or soft must set one at least, and hard >= soft.
func validRlimit(r *rlimit) error {
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
