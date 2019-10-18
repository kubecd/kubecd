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
	"github.com/spf13/cobra"
	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/model"
)

var (
	renderReleases     []string
	renderCluster      string
	renderInit         bool
	renderGitlab       bool
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "show helm templates and plain kubernetes YAML resources",
	Long:  ``,
	Args:  clusterFlagOrEnvArg(&renderCluster),
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		envsToApply, err := environmentsToApply(kcdConfig, args)
		if err != nil {
			return err
		}
		commandsToRun, err := commandsToRender(envsToApply)
		if err != nil {
			return err
		}
		for _, argv := range commandsToRun {
			if err = runCommand(false, true, argv); err != nil {
				return err
			}
		}
		return nil
	},
}

func commandsToRender(envsToApply []*model.Environment) ([][]string, error) {
	commandsToRun := make([][]string, 0)
	for _, env := range envsToApply {
		if renderInit {
			initCmds, err := commandsToInit(envsToApply, renderGitlab)
			if err != nil {
				return nil, err
			}
			for _, cmd := range initCmds {
				commandsToRun = append(commandsToRun, cmd)
			}
		}
		deployCmds, err := helm.TemplateCommands(env, renderReleases)
		if err != nil {
			return nil, err
		}
		for _, cmd := range deployCmds {
			commandsToRun = append(commandsToRun, cmd)
		}
	}
	return commandsToRun, nil
}

func init() {
	rootCmd.AddCommand(renderCmd)
	renderCmd.Flags().StringSliceVarP(&renderReleases, "releases", "r", []string{}, "generate template only these releases")
	renderCmd.Flags().StringVarP(&renderCluster, "cluster", "c", "", "template all environments in CLUSTER")
	renderCmd.Flags().BoolVar(&renderInit, "init", false, "initialize credentials and contexts")
	renderCmd.Flags().BoolVar(&renderGitlab, "gitlab", false, "initialize in gitlab mode")
}
