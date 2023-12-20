// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/google/gops/agent"
	"github.com/grafana/pyroscope-go"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/klog/v2"
)

// DaemonMain runs controllerContext handlers.
func DaemonMain() {

	// Print version info for debug.
	if len(controllerContext.Cfg.CommitVersion) > 0 {
		klog.Infof("CommitVersion: %v", controllerContext.Cfg.CommitVersion)
	}
	if len(controllerContext.Cfg.CommitTime) > 0 {
		klog.Infof("CommitTime: %v", controllerContext.Cfg.CommitTime)
	}
	if len(controllerContext.Cfg.AppVersion) > 0 {
		klog.Infof("AppVersion: %v", controllerContext.Cfg.AppVersion)
	}

	// Set golang max procs.
	currentP := runtime.GOMAXPROCS(-1)
	klog.Infof("Default max golang procs: %d", currentP)
	if currentP > int(controllerContext.Cfg.GoMaxProcs) {
		p := runtime.GOMAXPROCS(int(controllerContext.Cfg.GoMaxProcs))
		klog.Infof("Change max golang procs to %d", p)
	}

	// Load global Comfigmap.
	if err := controllerContext.LoadConfigmap(); err != nil {
		klog.Warning(err)
	}
	klog.Infof("Eni-controller config: %+v", controllerContext.Cfg)

	// Set up gops.
	if controllerContext.Cfg.GopsListenPort != "" {
		address := "127.0.0.1:" + controllerContext.Cfg.GopsListenPort
		op := agent.Options{
			ShutdownCleanup: true,
			Addr:            address,
		}
		if err := agent.Listen(op); err != nil {
			klog.Fatalf("gops failed to listen on %s: %v", address, err)
		}
		defer agent.Close()
		klog.Infof("gops is listkcrowng on %s", address)
	}

	// Set up pyroscope.
	if controllerContext.Cfg.PyroscopeAddress != "" {
		klog.Infof("pyroscope works in push mode with server: %s", controllerContext.Cfg.PyroscopeAddress)
		node, e := os.Hostname()
		if e != nil || len(node) == 0 {
			klog.Fatalf("Failed to get hostname: %v", e)
		}
		_, e = pyroscope.Start(pyroscope.Config{
			ApplicationName: binNameController,
			ServerAddress:   controllerContext.Cfg.PyroscopeAddress,
			Logger:          nil,
			Tags:            map[string]string{"node": node},
			ProfileTypes: []pyroscope.ProfileType{
				pyroscope.ProfileCPU,
				pyroscope.ProfileAllocObjects,
				pyroscope.ProfileAllocSpace,
				pyroscope.ProfileInuseObjects,
				pyroscope.ProfileInuseSpace,
			},
		})
		if e != nil {
			klog.Fatalf("Failed to setup pyroscope: %v", e)
		}
	}

	controllerContext.InnerCtx, controllerContext.InnerCancel = context.WithCancel(context.Background())
	klog.Info("Begin to initialize controller metrics HTTP server")

	klog.Info("Begin to initialize controller runtime manager")
	mgr, err := newCRDManager(&controllerContext.Cfg)
	if nil != err {
		klog.Fatal(err.Error())
	}
	controllerContext.CRDManager = mgr

	clientSet, err := initK8sClientSet()
	if nil != err {
		klog.Fatal(err.Error())
	}
	controllerContext.ClientSet = clientSet

	dynamicClient, err := initDynamicClient()
	if nil != err {
		klog.Fatal(err.Error())
	}
	controllerContext.DynamicClient = dynamicClient

	// init managers...
	initControllerServiceManagers(controllerContext)

	go func() {
		klog.Info("Starting controller runtime manager")
		if err := mgr.Start(controllerContext.InnerCtx); err != nil {
			klog.Fatal(err.Error())
		}
	}()
	waitForCacheSync := mgr.GetCache().WaitForCacheSync(controllerContext.InnerCtx)
	if !waitForCacheSync {
		klog.Fatal("failed to wait for syncing controller-runtime cache")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	WatchSignal(sigCh)
}

// WatchSignal notifies the signal to shut down controllerContext handlers.
func WatchSignal(sigCh chan os.Signal) {
	for sig := range sigCh {
		klog.Warningf("received shutdown", "signal", sig)

		// Cancel the internal context of controller.
		if controllerContext.InnerCancel != nil {
			controllerContext.InnerCancel()
		}
		// others...
	}
}

func initControllerServiceManagers(ctrlctx *ControllerContext) {

}

// initK8sClientSet will new kubernetes Clientset
func initK8sClientSet() (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if nil != err {
		return nil, fmt.Errorf("failed to init K8s clientset: %v", err)
	}

	return clientSet, nil
}

func initDynamicClient() (*dynamic.DynamicClient, error) {
	dynamicClient, err := dynamic.NewForConfig(ctrl.GetConfigOrDie())
	if nil != err {
		return nil, fmt.Errorf("failed to init Kubernetes dynamic client: %v", err)
	}

	return dynamicClient, nil
}
