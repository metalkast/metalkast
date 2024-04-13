package main

import (
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/metalkast/metalkast/cmd/kast/bootstrap"
	"github.com/metalkast/metalkast/cmd/kast/log"
	"github.com/spf13/cobra"
	clusterctllog "sigs.k8s.io/cluster-api/cmd/clusterctl/log"
)

// bootstrapCmd represents the bootstrap command
var (
	bootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstraps a metalkast cluster from the provided manifests",
		Long: `Bootstraps a metalkast cluster from the provided manifests.

On a high level the bootstrap process performs the following steps:

1. Boot the bootstrap node from a live CD and start and initiate a Kubernetes cluster
2. Provision target cluster from the bootstrap node cluster
3. Pivot: Move the target cluster's manifests to itself
4. Join the rest of the nodes to the target cluster

Usage:

kast bootstrap MANIFESTS...
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runDirectory, err := CreateRunDirectory()
			if err != nil {
				log.Log.Error(err, "Failed to create run directory")
				return err
			}
			cliLogger, err := log.NewLogger(log.LoggerOptions{
				OutputPath: path.Join(runDirectory, "run.log"),
			})
			if err != nil {
				return err
			}
			defer (cliLogger.GetSink()).(*log.TeaLogSink).Close()
			log.SetLogger(cliLogger)
			clusterctllog.SetLogger(cliLogger.V(1).WithName("clusterctl"))

			err = checkIpmitoolInstalled()
			if err != nil {
				switch runtime.GOOS {
				case "linux":
					log.Log.V(1).Info("To install ipmitool on Ubuntu run:\n\n\tapt-get install -y ipmitool")
				case "darwin":
					log.Log.V(1).Info("To install ipmitool on MacOS run:\n\n\tbrew install ipmitool")
				default:
					log.Log.Error(fmt.Errorf("platform %s is not suported", runtime.GOOS), "kast is not supported on your platform")
				}
				return fmt.Errorf("failed to detect ipmitool installation: %w", err)
			}

			b, err := bootstrap.FromManifests(args)
			if err != nil {
				return fmt.Errorf("failed to configure bootstrap from manifests: %w", err)
			}

			if err := b.Run(bootstrap.BootstrapOptions{
				BootstrapNodeOptions: bootstrap.BootstrapNodeOptions{
					KubeCfgDestPath: path.Join(runDirectory, "bootstrap.kubeconfig"),
					SSHKeyDestPath:  path.Join(runDirectory, "ssh.key"),
				},
			}); err != nil {
				log.Log.Error(err, "Bootstrap failed")
				return err
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}

func checkIpmitoolInstalled() error {
	cmd := exec.Command("ipmitool", "-V")
	cmd.Stderr = nil
	cmd.Stdout = nil
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run ipmitool: %w", err)
	}
	const expectedVersionPrefix = "ipmitool version 1."
	if !strings.HasPrefix(string(out), expectedVersionPrefix) {
		return fmt.Errorf(
			"failed to detect ipmitool version: version reported '%s', but expected '1.0+'",
			strings.TrimSpace(strings.TrimPrefix(string(out), expectedVersionPrefix)),
		)
	}
	return nil
}
