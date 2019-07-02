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
	"os"
	"path"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "print shell completion script",
	Long: `To load completion, run 

. <(kcd completion)

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		switch path.Base(os.Getenv("SHELL")) {
		case "zsh":
			err = rootCmd.GenZshCompletion(os.Stdout)
		case "pwsh", "pwsh.exe", "powershell.exe":
			err = rootCmd.GenPowerShellCompletion(os.Stdout)
		default:
			err = rootCmd.GenBashCompletion(os.Stdout)
		}
		if err != nil {
			return fmt.Errorf("failed generating completion: %v", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// completionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// completionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
