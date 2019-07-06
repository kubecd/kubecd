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
	"github.com/zedge/kubecd/pkg/kubecd"
	"github.com/zedge/kubecd/pkg/model"

	"github.com/spf13/cobra"
)

var (
	pollPatch    bool
	pollReleases []string
	pollImage    string
	pollCluster  string
)

// pollCmd represents the poll command
var pollCmd = &cobra.Command{
	Use:   "poll",
	Short: "poll for new images in registries",
	Long:  ``,
	Args:  matchAll(cobra.RangeArgs(0, 1), clusterFlagOrEnvArg(&pollCluster)),
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		releaseFilters := makePollReleaseFilters(cmd, args)
		imageIndex, err := kubecd.ImageReleaseIndex(kcdConfig, releaseFilters...)
		if err != nil {
			return err
		}
		imageTags, err := kubecd.BuildTagIndexFromDockerRegistries(imageIndex)
		if err != nil {
			return err
		}
		allUpdates := make([]kubecd.ImageUpdate, 0)
		for repo, releases := range imageIndex {
			fmt.Printf("image: %s\n", repo)
			for _, release := range releases {
				imageUpdates, err := kubecd.FindImageUpdatesForRelease(release, imageTags)
				if err != nil {
					return err
				}
				allUpdates = append(allUpdates, imageUpdates...)
			}
		}
		if len(allUpdates) == 0 {
			fmt.Println("No updates found.")
			return nil
		}
		if err = patchReleasesFilesMaybe(allUpdates, pollPatch); err != nil {
			return err
		}
		return nil
	},
}

func makePollReleaseFilters(cmd *cobra.Command, args []string) []kubecd.ReleaseFilterFunc {
	filters := make([]kubecd.ReleaseFilterFunc, 0)
	if pollCluster != "" {
		filters = append(filters, kubecd.ClusterReleaseFilter(pollCluster))
	} else {
		filters = append(filters, kubecd.EnvironmentReleaseFilter(args[0]))
	}
	if len(pollReleases) > 0 {
		filters = append(filters, kubecd.ReleaseFilter(pollReleases))
	}
	if pollImage != "" {
		filters = append(filters, kubecd.ImageReleaseFilter(pollImage))
	}
	return filters
}

func init() {
	rootCmd.AddCommand(pollCmd)
	pollCmd.Flags().BoolVarP(&pollPatch, "patch", "p", false, "patch releases.yaml files with updated version")
	pollCmd.Flags().StringSliceVarP(&pollReleases, "releases", "r", []string{}, "poll one or more specific releases")
	pollCmd.Flags().StringVarP(&pollImage, "image", "i", "", "poll releases using this image")
	pollCmd.Flags().StringVarP(&pollCluster, "cluster", "c", "", "poll all releases in this cluster")
}
