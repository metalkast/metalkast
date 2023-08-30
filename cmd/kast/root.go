package main

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "kast",
	Short:        "Quickly provision Kubernetes bare metal clusters",
	Long:         `Kast is a tool for quickly provisioning Kubernetes bare metal clusters`,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	initLoggers()
}

func initLoggers() {
	// Inspired from https://github.com/fluxcd/flux2/pull/3932
	ctrllog.SetLogger(logr.New(ctrllog.NullLogSink{}))
}
