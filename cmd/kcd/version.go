package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "snapshot"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show KubeCD version and exit",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
