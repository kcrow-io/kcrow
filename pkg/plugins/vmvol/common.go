package plugins

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func commonOpt(spec *corev1.PersistentVolumeSpec) []string {
	var (
		ret []string
	)
	if spec.MountOptions != nil {
		return append(ret, "-o", strings.Join(spec.MountOptions, ","))
	}
	return ret
}
