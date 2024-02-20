// Copyright 2023 Authors of kcrow
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var binNameController = filepath.Base(os.Args[0])

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   binNameController,
	Short: binNameController,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}
