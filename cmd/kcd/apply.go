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
package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/model"
)

var applyReleases []string
var applyCluster string
var applyInit bool
var applyGitlab bool
var applyDryRun bool
var applyDebug bool

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply changes to Kubernetes",
	Long:  ``,
	Args:  clusterFlagOrEnvArg(&applyCluster),
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
			if err = runCommand(false, false, argv); err != nil {
				return err
			}
		}
		return nil
	},
}

func commandsToApply(envsToApply []*model.Environment) ([][]string, error) {
	commandsToRun := make([][]string, 0)
	for _, env := range envsToApply {
		if applyInit {
			initCmds, err := commandsToInit(envsToApply, applyGitlab)
			if err != nil {
				return nil, err
			}
			for _, cmd := range initCmds {
				commandsToRun = append(commandsToRun, cmd)
			}
		}
		deployCmds, err := helm.DeployCommands(env, applyDryRun, applyDebug, applyReleases)
		if err != nil {
			return nil, err
		}
		for _, cmd := range deployCmds {
			commandsToRun = append(commandsToRun, cmd)
		}
	}
	return commandsToRun, nil
}

func environmentsFromArgs(kcdConfig *model.KubeCDConfig, cluster string, args []string) ([]*model.Environment, error) {
	if len(args) > 0 {
		for _, envName := range args {
			env := kcdConfig.GetEnvironment(envName)
			if env == nil {
				return nil, fmt.Errorf(`unknown environment: %q`, envName)
			}
			return []*model.Environment{env}, nil
		}
	}
	if cluster == "" {
		return nil, errors.New(`specify --cluster flag or ENV arg`)
	}
	if !kcdConfig.HasCluster(cluster) {
		return nil, fmt.Errorf(`unknown cluster: %q`, cluster)
	}
	return kcdConfig.GetEnvironmentsInCluster(cluster), nil
}

func environmentsToApply(kcdConfig *model.KubeCDConfig, args []string) ([]*model.Environment, error) {
	return environmentsFromArgs(kcdConfig, applyCluster, args)
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVarP(&applyDryRun, "dry-run", "n", false, "dry run mode, only print commands")
	applyCmd.Flags().BoolVar(&applyDebug, "debug", false, "run helm with --debug")
	applyCmd.Flags().StringSliceVarP(&applyReleases, "releases", "r", []string{}, "apply only these releases")
	applyCmd.Flags().StringVarP(&applyCluster, "cluster", "c", "", "apply all environments in CLUSTER")
	applyCmd.Flags().BoolVar(&applyInit, "init", false, "initialize credentials and contexts")
	applyCmd.Flags().BoolVar(&applyGitlab, "gitlab", false, "initialize in gitlab mode")
}
