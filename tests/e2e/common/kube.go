package common

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	kubernetes.Interface
	client.Client

	Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface

	RestConfig() *rest.Config
}

type k8sclient struct {
	kubernetes.Interface
	client.Client
	*dynamic.DynamicClient

	cfg *rest.Config
}

func (k *k8sclient) RestConfig() *rest.Config {
	return k.cfg
}

func (k *k8sclient) CreateOrUpdate(ctx context.Context, obj client.Object) error {

	gvk := obj.GetObjectKind()
	err := k.Client.Get(ctx, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return k.Client.Create(ctx, obj)
		} else {
			return fmt.Errorf("failed to get %s '%s': %w", gvk, obj.GetName(), err)
		}
	} else {
		return k.Client.Update(ctx, obj)
	}
}

// initK8sClientSet 函数用于初始化 Kubernetes 客户端集合
func initK8sClientSet(cfg *rest.Config) (kubernetes.Interface, error) {

	clientSet, err := kubernetes.NewForConfig(cfg)
	if nil != err {
		return nil, fmt.Errorf("failed to init K8s clientset: %v", err)
	}
	return clientSet, nil
}

// initDynamicClient
func initDynamicClient(cfg *rest.Config) (*dynamic.DynamicClient, error) {

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if nil != err {
		return nil, fmt.Errorf("failed to init Kubernetes dynamic client: %v", err)
	}
	return dynamicClient, nil
}

// initRuntimeClient
func initRuntimeClient(cfg *rest.Config) (client.Client, error) {

	runtimeClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to init Kubernetes runtime client: %v", err)
	}
	return runtimeClient, nil
}

func MustGetClient() *k8sclient {
	cli, err := GetClient()
	if err != nil {
		panic(err)
	}
	return cli
}

func GetClient() (*k8sclient, error) {
	var config = ctrl.GetConfigOrDie()
	dynamicClient, err := initDynamicClient(config)
	if err != nil {
		return nil, err
	}
	clientset, err := initK8sClientSet(config)
	if err != nil {
		return nil, err
	}

	runtimeClient, err := initRuntimeClient(config)
	if err != nil {
		return nil, err
	}

	return &k8sclient{
		cfg:           config,
		Interface:     clientset,
		Client:        runtimeClient,
		DynamicClient: dynamicClient,
	}, nil

}
