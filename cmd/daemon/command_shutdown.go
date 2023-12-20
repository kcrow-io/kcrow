// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// shutdownCmd represents the shutdown command.
var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "shutdown " + binNameController,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO (Icarus9913)
		klog.Infof("Shutdown %s...", binNameController)
	},
}

func init() {
	rootCmd.AddCommand(shutdownCmd)
}
