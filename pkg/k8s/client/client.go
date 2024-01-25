package client

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type Client struct {
	KubeClient kubernetes.Interface
}

func New(kubeConfigFile string) (*Client, error) {
	var (
		cfg *rest.Config
		cli = &Client{}
		err error
	)

	if kubeConfigFile == "" {
		klog.Infof("no --kubeconfig, use in-cluster kubernetes config")
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	}
	if err != nil {
		klog.Errorf("failed to build kubeconfig %v", err)
		return nil, err
	}

	cfg.QPS = 1000
	cfg.Burst = 2000
	cfg.Timeout = 30 * time.Second

	cfg.ContentType = "application/vnd.kubernetes.protobuf"
	cfg.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("init kubernetes client failed %v", err)
		return nil, err
	}
	cli.KubeClient = kubeClient

	return cli, nil
}
