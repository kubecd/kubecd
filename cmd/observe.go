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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/kubecd"
	"github.com/zedge/kubecd/pkg/model"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var observeImage string
var observeChart string
var observePatch bool

// observeCmd represents the observe command
var observeCmd = &cobra.Command{
	Use:   "observe",
	Short: "observe a new version of an image or chart",
	Long: ``,
	Args: matchAll(cobra.RangeArgs(0, 1), imageOrChart(&observeImage, &observeChart)),
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		if observeImage != "" {
			return observeImageTag(kcdConfig, cmd, args)
		}
		return observeChartVersion(kcdConfig, cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(observeCmd)
	observeCmd.Flags().StringVarP(&observeImage, "image", "i", "", "a new image, including tag")
	observeCmd.Flags().StringVar(&observeChart, "chart", "", "a new chart version")
	observeCmd.Flags().BoolVar(&observePatch, "patch", false, "patch release files with updated tags")
}

func observeImageTag(kcdConfig *model.KubeCDConfig, cmd *cobra.Command, args []string) error {
	imageIndex, err := imageReleaseIndex(kcdConfig)
	if err != nil {
		return errors.Wrapf(err, `while observing --image %q`, observeImage)
	}
	verb := "May"
	if observePatch {
		verb = "Will"
	}
	newImage := kubecd.NewDockerImageRef(observeImage)
	updateFound := false
	for img, releases := range imageIndex {
		for _, release := range releases {
			updates, err := kubecd.ReleaseWantsImageUpdate(release, newImage)
			if err != nil {
				return errors.Wrapf(err, `release %q: while looking for image updates for %q`, release.Name, img)
			}
			for _, update := range updates {
				fmt.Printf("%s upgrade release %q image %q to %q because %s\n", verb, release.Name, update.ImageRepo, update.NewTag, update.Reason)
				updateFound = true
			}
			if observePatch {
				if err = patchImageUpdates(release.FromFile, updates); err != nil {
					return err
				}
			}
		}
	}
	if !updateFound {
		fmt.Printf("No matching release found for image %s.\n", observeImage)
	}
	return nil
}

func patchImageUpdates(releasesFile string, updates []kubecd.ImageUpdate) error {
	var modYaml interface{}
	data, err := ioutil.ReadFile(releasesFile)
	if err != nil {
		return errors.Wrapf(err, `error reading %q`, releasesFile)
	}
	err = yaml.Unmarshal(data, &modYaml)
	if err != nil {
		return errors.Wrapf(err, `error decoding yaml in %q`, releasesFile)
	}
	tmpReleases := modYaml.(map[interface{}]interface{})["releases"]
	if tmpReleases == nil {
		return fmt.Errorf(`%s: no "releases" found`, releasesFile)
	}
	releases, ok := tmpReleases.([]interface{})
	if !ok {
		return fmt.Errorf(`%s: "releases" not a list`, releasesFile)
	}
	for _, tmpRel := range releases {
		modRel := tmpRel.(map[interface{}]interface{})
		relName := modRel["name"].(string)
		for _, update := range updates {
			if relName != update.Release.Name {
				continue
			}
			if _, found := modRel["values"]; !found {
				modRel["values"] = make(map[interface{}]interface{})
			}
			foundVal := false
			values := modRel["values"].([]interface{})
			for _, tmpVal := range values {
				modVal := tmpVal.(map[string]interface{})
				if modVal["key"].(string) != update.TagValue {
					continue
				}
				modVal["value"] = update.NewTag
				foundVal = true
				break
			}
			if !foundVal {
				values = append(values, map[interface{}]interface{}{"key": update.TagValue, "value": update.NewTag})
			}
		}
	}
	if err = writeIndentedYamlToFile(releasesFile, modYaml); err != nil {
		return err
	}
	return nil
}

func observeChartVersion(kcdConfig *model.KubeCDConfig, cmd *cobra.Command, args []string) error {
	return nil
}

func imageOrChart(image, chart *string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if (*chart != "" && *image != "") || (*chart == "" && *image == "") {
			return errors.New("must specify --image or --chart")
		}
		return nil
	}
}

func imageReleaseIndex(kcdConfig *model.KubeCDConfig) (map[string][]*model.Release, error) {
	result := make(map[string][]*model.Release)
	for _, release := range kcdConfig.AllReleases() {
		//fmt.Printf("evaluating release %q\n", release.Name)
		values, err := helm.GetResolvedValues(release)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving values for env %q release %q", release.Environment.Name, release.Name)
		}
		//fmt.Printf("release %q triggers: %#v\n", release.Name, release.Triggers)
		for _, t := range release.Triggers {
			if t.Image == nil {
				//fmt.Printf("release %q has no trigger\n", release.Name)
				continue
			}
			repo := helm. GetImageRepoFromImageTrigger(t.Image, values)
			//fmt.Printf("release %q repo: %q\n", release.Name, repo)
			if _, found := result[repo]; !found {
				result[repo] = make([]*model.Release, 0)
			}
			result[repo] = append(result[repo], release)
		}
	}
	return result, nil
}
