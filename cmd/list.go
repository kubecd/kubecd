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
	"github.com/spf13/cobra"
	"github.com/zedge/kubecd/pkg/model"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list {env,release,cluster}",
	Short: "list clusters, environments or releases",
	Long: ``,
	Args: matchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"env", "envs", "release", "releases", "cluster", "clusters"},
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		switch args[0] {
		case "env", "envs":
			for _, env := range kcdConfig.Environments {
				fmt.Println(env.Name)
			}
		case "release", "releases":
			for _, env := range kcdConfig.Environments {
				for _, release := range env.AllReleases() {
					fmt.Printf("%s -r %s\n", env.Name, release.Name)
				}
			}
		case "cluster", "clusters":
			for _, cluster := range kcdConfig.AllClusters() {
				fmt.Println(cluster.Name)
			}
		}
		return nil
	},
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

func init() {
	rootCmd.AddCommand(listCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
