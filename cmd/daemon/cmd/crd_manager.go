// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func newCRDManager(cfg *Config) (cluster.Cluster, error) {

	config := ctrl.GetConfigOrDie()
	config.Burst = 200
	config.QPS = 100

	cacheopt := cache.Options{
		Scheme: scheme,
	}

	clus, err := cluster.New(config, func(o *cluster.Options) {
		o.Cache = cacheopt
		o.Scheme = scheme
		o.Logger = logr.New(log.NullLogSink{})
	})

	if err != nil {
		return nil, err
	}

	return clus, nil
}

func TransNode(in interface{}) (out interface{}, err error) {
	v, ok := in.(*corev1.Node)
	if ok {
		return &corev1.Node{
			TypeMeta:   v.TypeMeta,
			ObjectMeta: v.ObjectMeta,
			Spec:       *v.Spec.DeepCopy(),
		}, nil
	}
	return nil, fmt.Errorf("it is not node type")
}
