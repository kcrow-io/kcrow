package common

import (
	"flag"

	_ "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	DockerMirror string
)

func init() {
	flag.StringVar(&DockerMirror, "mirror", "docker.io", "set docker hub mirror")
}
