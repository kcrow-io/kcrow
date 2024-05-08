package common

import "flag"

var (
	kubeconfig string
)

const (
	DefaultKubeconfig = "/etc/kubeconfig/confit.yaml"
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", DefaultKubeconfig, "The path of the kubeconfig file")
}
