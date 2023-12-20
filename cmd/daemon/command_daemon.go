// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"runtime/debug"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: binNameController + " daemon",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if e := recover(); nil != e {
				klog.Errorf("Panic details: %v", e)
				debug.PrintStack()
			}
		}()

		DaemonMain()
	},
}

func init() {
	controllerContext.BindControllerDaemonFlags(daemonCmd.PersistentFlags())
	if err := ParseConfiguration(); nil != err {
		klog.Exitf("Failed to register ENV for kcrow-controller: %v", err)
	}
	controllerContext.Verify()

	rootCmd.AddCommand(daemonCmd)
}
