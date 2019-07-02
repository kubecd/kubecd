/*
Copyright © 2019 Zedge, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/mitchellh/colorstring"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/model"
	"github.com/zedge/kubecd/pkg/provider"
	"os"
	"os/exec"
	"strings"
)

var applyReleases []string
var applyCluster string
var applyInit bool
var applyGitlab bool

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply changes to Kubernetes",
	Long: ``,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		envsToApply, err := environmentsToApply(kcdConfig, args)
		if err != nil {
			return err
		}
		commandsToRun, err := commandsToApply(envsToApply)
		if err != nil {
			return err
		}
		for _, argv := range commandsToRun {
			printCmd := strings.Join(argv, " ")
			_, _ = colorstring.Printf("[yellow]%s\n", printCmd)
			cmd := exec.Command(argv[0], argv[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				return errors.Wrapf(err, "command failed: %q", printCmd)
			}
		}
		return nil
	},
}

func commandsToApply(envsToApply []*model.Environment) ([][]string, error) {
	commandsToRun := make([][]string, 0)
	clusterInitialized := make(map[string]bool)
	for _, env := range envsToApply {
		cluster := env.GetCluster()
		cp, err := provider.GetClusterProvider(cluster, applyGitlab)
		if err != nil {
			return nil, err
		}
		if applyInit {
			if _, initialized := clusterInitialized[cluster.Name]; !initialized {
				cmds, err := cp.GetClusterInitCommands()
				if err != nil {
					return nil, err
				}
				for _, cmd := range cmds {
					commandsToRun = append(commandsToRun, cmd)
				}
			}
			clusterInitialized[cluster.Name] = true
			for _, cmd := range provider.GetContextInitCommands(cp, env) {
				commandsToRun = append(commandsToRun, cmd)
			}
		}
		deployCmds, err := helm.DeployCommands(env, dryRun, debug, applyReleases)
		if err != nil {
			return nil, err
		}
		for _, cmd := range deployCmds {
			commandsToRun = append(commandsToRun, cmd)
		}
	}
	return commandsToRun, nil
}

func environmentsToApply(kcdConfig *model.KubeCDConfig, args []string) ([]*model.Environment, error) {
	if len(args) > 0 {
		for _, envName := range args {
			env := kcdConfig.GetEnvironment(envName)
			if env == nil {
				return nil, fmt.Errorf(`unknown environment: %q`, envName)
			}
			return []*model.Environment{env}, nil
		}
	}
	if applyCluster == "" {
		return nil, errors.New(`must apply either --cluster or [env...]`)
	}
	if !kcdConfig.HasCluster(applyCluster) {
		return nil, fmt.Errorf(`unknown cluster: %q`, applyCluster)
	}
	return kcdConfig.GetEnvironmentsInCluster(applyCluster), nil
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "dry run mode, only print commands")
	applyCmd.Flags().BoolVar(&debug, "debug", false, "run helm with --debug")
	applyCmd.Flags().StringSliceVarP(&applyReleases, "releases", "r", []string{}, "apply only these releases")
	applyCmd.Flags().StringVarP(&applyCluster, "cluster", "c", "", "apply all environments in CLUSTER")
	applyCmd.Flags().BoolVar(&applyInit, "init", false, "initialize credentials and contexts")
	applyCmd.Flags().BoolVar(&applyGitlab, "gitlab", false, "initialize in gitlab mode")
}
