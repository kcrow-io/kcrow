package util

import (
	"strings"

	"k8s.io/klog/v2"
)

func GetValueFromEnvByKey(data []string, key string) string {
	for _, envLine := range data {
		words := strings.SplitN(envLine, "=", 2)
		if len(words) != 2 {
			klog.Warningf("environment error for %s", envLine)
			continue
		}

		if words[0] == key {
			return words[1]
		}
	}
	return ""
}

func IterEnvVar(data []string, fn func(k, v string) bool) {
	for _, line := range data {
		words := strings.SplitN(line, "=", 2)
		if len(words) != 2 {
			klog.Warningf("environment error for %s", line)
			continue
		}
		ok := fn(words[0], words[1])
		if !ok {
			return
		}
	}
}
