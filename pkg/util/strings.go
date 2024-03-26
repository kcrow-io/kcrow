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
