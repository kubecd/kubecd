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
	"github.com/spf13/cobra"
)

var (
	diffDryRun   bool
	diffReleases []string
	diffCluster  string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "show diff between deployed resources and local config",
	Long:  `Show a colored unified diff of the deployed resources and the config in the current dir.`,
	Args:  clusterFlagOrEnvArg(&diffCluster),
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		envsToDiff, err := environmentsFromArgs(kcdConfig, diffCluster, args)
		if err != nil {
			return err
		}
		commandsToRun, err := commandsToDiff(envsToDiff)
		if err != nil {
			return err
		}
		for _, argv := range commandsToRun {
			if err = runCommand(diffDryRun, false, argv); err != nil {
				return err
			}
		}
		return nil
	},
}

func commandsToDiff(envsToDiff []*model.Environment) ([][]string, error) {
	commandsToRun := make([][]string, 0)
	for _, env := range envsToDiff {
		diffCmds, err := helm.DiffCommands(env, diffReleases)
		if err != nil {
			return nil, err
		}
		commandsToRun = append(commandsToRun, diffCmds...)
	}
	return commandsToRun, nil
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolVarP(&diffDryRun, "dry-run", "n", false, "dry run mode, only print commands")
	diffCmd.Flags().StringSliceVarP(&diffReleases, "releases", "r", []string{}, "diff only these releases")
	diffCmd.Flags().StringVarP(&diffCluster, "cluster", "c", "", "diff all environments in CLUSTER")
}
