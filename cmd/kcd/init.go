/*
 * Copyright 2018-2019 Zedge, Inc.
 * Copyright 2019-2020 Stig Sæther Nordahl Bakken
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package main

import (
	"github.com/kubecd/kubecd/pkg/helm"
	"github.com/kubecd/kubecd/pkg/model"
	"github.com/kubecd/kubecd/pkg/provider"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var initCluster string
var initContextsOnly bool
var initDryRun bool
var initGitlabMode bool

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [ENV]",
	Short: "Initialize credentials and contexts",
	Long:  ``,
	Args:  clusterFlagOrEnvArg(&initCluster),
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		cmds := helm.RepoSetupCommands(kcdConfig.HelmRepos)
		envsToInit, err := environmentsFromArgs(kcdConfig, initCluster, args)
		if err != nil {
			return err
		}
		initCmds, err := commandsToInit(envsToInit, initGitlabMode)
		if err != nil {
			return err
		}
		cmds = append(cmds, initCmds...)
		for _, argv := range cmds {
			if err = runCommand(initDryRun, false, argv); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initCluster, "cluster", "", "Initialize contexts for all environments in a cluster")
	initCmd.Flags().BoolVar(&initContextsOnly, "contexts-only", false, "initialize contexts only, assuming that cluster credentials are set up")
	initCmd.Flags().BoolVarP(&initDryRun, "dry-run", "n", false, "print commands instead of running them")
	initCmd.Flags().BoolVar(&initGitlabMode, "gitlab", false, "grab kube config from GitLab environment")
}

func commandsToInit(envs []*model.Environment, gitlabMode bool) ([][]string, error) {
	commands := make([][]string, 0)
	clusterInitialized := make(map[string]bool)
	for _, env := range envs {
		cluster := env.GetCluster()
		cp, err := provider.GetClusterProvider(cluster, gitlabMode)
		if err != nil {
			return nil, err
		}
		if _, initialized := clusterInitialized[cluster.Name]; !initialized {
			cmds, err := cp.GetClusterInitCommands()
			if err != nil {
				return nil, err
			}
			commands = append(commands, cmds...)
		}
		clusterInitialized[cluster.Name] = true
		commands = append(commands, provider.GetContextInitCommands(cp, env)...)
	}
	return commands, nil
}

func clusterFlagOrEnvArg(clusterFlag *string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if *clusterFlag == "" && len(args) != 1 {
			return errors.New("specify --cluster flag or ENV arg")
		}
		return nil
	}
}

func matchAll(checks ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, check := range checks {
			if err := check(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
