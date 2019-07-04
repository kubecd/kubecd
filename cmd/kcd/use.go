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
	"github.com/spf13/cobra"
	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/model"
)

var useDryRun bool

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use {ENV}",
	Short: "switch kube context to the specified environment",
	Long: ``,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		env := kcdConfig.GetEnvironment(args[0])
		if env == nil {
			return fmt.Errorf(`unknown environment %q`, args[0])
		}
		argv := helm.UseContextCommand(env.Name)
		if err = runCommand(useDryRun, argv); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.Flags().BoolVarP(&useDryRun, "dry-run", "n", false, "print commands instead of running them")
}
