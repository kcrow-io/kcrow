package common

import "flag"

var (
	kubeconfig string
)

const (
	DefaultKubeconfig = "/etc/kubeconfig/config.yaml"
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", DefaultKubeconfig, "The path of the kubeconfig file")
}
