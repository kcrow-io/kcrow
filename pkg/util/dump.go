package util

import (
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// Dump one or more objects, with an optional global prefix and per-object tags.
func Dump(args ...interface{}) {
	var (
		prefix string
		idx    int
	)

	if len(args)&0x1 == 1 {
		prefix = args[0].(string)
		idx++
	}

	for ; idx < len(args)-1; idx += 2 {
		tag, obj := args[idx], args[idx+1]
		msg, err := yaml.Marshal(obj)
		if err != nil {
			klog.Infof("%s: %s: failed to dump object: %v", prefix, tag, err)
			continue
		}

		if prefix != "" {
			klog.Infof("%s: %s:", prefix, tag)
			for _, line := range strings.Split(strings.TrimSpace(string(msg)), "\n") {
				klog.Infof("%s:    %s", prefix, line)
			}
		} else {
			klog.Infof("%s:", tag)
			for _, line := range strings.Split(strings.TrimSpace(string(msg)), "\n") {
				klog.Infof("  %s", line)
			}
		}
	}
}
