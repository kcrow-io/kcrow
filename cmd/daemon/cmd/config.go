// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

var controllerContext = new(ControllerContext)

type envConf struct {
	envName          string
	defaultValue     string
	required         bool
	associateStrKey  *string
	associateBoolKey *bool
	associateIntKey  *int
}

// EnvInfo collects the env and relevant agentContext properties.
var envInfo = []envConf{
	{"GIT_COMMIT_VERSION", "", false, &controllerContext.Cfg.CommitVersion, nil, nil},
	{"GIT_COMMIT_TIME", "", false, &controllerContext.Cfg.CommitTime, nil, nil},
	{"VERSION", "", false, &controllerContext.Cfg.AppVersion, nil, nil},
	{"GOLANG_ENV_MAXPROCS", "8", false, nil, nil, &controllerContext.Cfg.GoMaxProcs},

	// environ also used in nri library
	{"KCROW_NRIPLUGIN_SOCKET", "/var/run/nri/nri.sock", false, &controllerContext.Cfg.NriSockPath, nil, nil},
}

type Config struct {
	CommitVersion string
	CommitTime    string
	AppVersion    string
	GoMaxProcs    int
	NriSockPath   string

	// flags
	GopsListenPort   string
	PyroscopeAddress string
}

type ControllerContext struct {
	Cfg Config

	// InnerCtx is the context that can be used during shutdown.
	// It will be cancelled after receiving an interrupt or termination signal.
	InnerCtx    context.Context
	InnerCancel context.CancelFunc

	// cluster
	CRDCluster cluster.Cluster

	// probe
	IsStartupProbe atomic.Bool
}

// BindControllerDaemonFlags bind controller cli daemon flags
func (cc *ControllerContext) BindControllerDaemonFlags(fs *pflag.FlagSet) {
	var (
		kubeflags = &flag.FlagSet{}
		logflags  = &flag.FlagSet{}
	)

	klog.InitFlags(logflags)
	config.RegisterFlags(kubeflags)
	fs.AddGoFlagSet(logflags)
	fs.AddGoFlagSet(kubeflags)

	fs.StringVar(&cc.Cfg.GopsListenPort, "gops-port", "", "gops listen port")
	fs.StringVar(&cc.Cfg.PyroscopeAddress, "pyroscope-address", "", "pyroscope address")
}

// ParseConfiguration set the env to AgentConfiguration
func ParseConfiguration() error {
	var result string

	for i := range envInfo {
		env, ok := os.LookupEnv(envInfo[i].envName)
		if ok {
			result = strings.TrimSpace(env)
		} else {
			// if no env and required, set it to default value.
			result = envInfo[i].defaultValue
		}
		if len(result) == 0 {
			if envInfo[i].required {
				klog.Exitf(fmt.Sprintf("empty value of %s", envInfo[i].envName))
			} else {
				// if no env and none-required, just use the empty value.
				continue
			}
		}

		if envInfo[i].associateStrKey != nil {
			*(envInfo[i].associateStrKey) = result
		} else if envInfo[i].associateBoolKey != nil {
			b, err := strconv.ParseBool(result)
			if nil != err {
				return fmt.Errorf("error: %s require a bool value, but get %s", envInfo[i].envName, result)
			}
			*(envInfo[i].associateBoolKey) = b
		} else if envInfo[i].associateIntKey != nil {
			intVal, err := strconv.Atoi(result)
			if nil != err {
				return fmt.Errorf("error: %s require a int value, but get %s", envInfo[i].envName, result)
			}
			*(envInfo[i].associateIntKey) = intVal
		} else {
			return fmt.Errorf("error: %s doesn't match any controller context", envInfo[i].envName)
		}
	}

	return nil
}

// verify after retrieve all config
func (cc *ControllerContext) Verify() {
	// loglevel
}
