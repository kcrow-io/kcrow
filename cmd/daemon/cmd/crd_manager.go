// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/go-logr/logr"
	"github.com/kcrow-io/kcrow/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		ByObject: map[client.Object]cache.ByObject{
			&corev1.Node{}: {
				Transform: k8s.TransNode,
			},
			&corev1.Pod{}: {
				Transform: k8s.TransPod,
			},
		},
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
