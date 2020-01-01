/*
 * Copyright 2018-2020 Zedge, Inc.
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
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/kubecd/kubecd/pkg/image"
	"github.com/kubecd/kubecd/pkg/model"
	"github.com/kubecd/kubecd/pkg/updates"
)

var (
	observePatch    bool
	observeReleases []string
	observeImage    string
	observeChart    string
	observeVerify   bool
)

// observeCmd represents the observe command
var observeCmd = &cobra.Command{
	Use:   "observe [ENV]",
	Short: "observe a new version of an image or chart",
	Long:  ``,
	Args:  matchAll(cobra.RangeArgs(0, 1), imageOrChart(&observeImage, &observeChart)),
	RunE: func(cmd *cobra.Command, args []string) error {
		kcdConfig, err := model.NewConfigFromFile(environmentsFile)
		if err != nil {
			return err
		}
		if observeImage != "" {
			return observeImageTag(kcdConfig, cmd, args)
		}
		return observeChartVersion(kcdConfig)
	},
}

func init() {
	rootCmd.AddCommand(observeCmd)
	observeCmd.Flags().StringVarP(&observeImage, "image", "i", "", "a new image, including tag")
	observeCmd.Flags().StringSliceVarP(&observeReleases, "releases", "r", []string{}, "limit the update to or more specific releases")
	observeCmd.Flags().StringVar(&observeChart, "chart", "", "a new chart version")
	observeCmd.Flags().BoolVar(&observePatch, "patch", false, "patch release files with updated tags")
	observeCmd.Flags().BoolVar(&observeVerify, "verify", false, "verify that image:tag exists")
}

func observeVerifyImage(imageRepo string) error {
	imageRef := image.NewDockerImageRef(imageRepo)
	existingTags, err := image.GetTagsForDockerImage(imageRepo)
	if err != nil {
		return err
	}
	for _, tsTag := range existingTags {
		if tsTag.Tag == imageRef.Tag {
			return nil
		}
	}
	return fmt.Errorf(`tag %q not found for imageRepo %q`, imageRef.Tag, imageRef.WithoutTag())
}

func observeImageTag(kcdConfig *model.KubeCDConfig, cmd *cobra.Command, args []string) error {
	if observeVerify {
		if err := observeVerifyImage(observeImage); err != nil {
			return err
		}
	}
	releaseFilters := makeObserveReleaseFilters(args)
	imageIndex, err := updates.ImageReleaseIndex(kcdConfig, releaseFilters...)
	if err != nil {
		return err
	}
	newImage := image.NewDockerImageRef(observeImage)
	imageTags := updates.BuildTagIndexFromNewImageRef(newImage, imageIndex)
	allUpdates := make([]updates.ImageUpdate, 0)
	for _, release := range imageIndex[newImage.WithoutTag()] {
		imageUpdates, err := updates.FindImageUpdatesForRelease(release, imageTags)
		if err != nil {
			return err
		}
		allUpdates = append(allUpdates, imageUpdates...)
	}
	if len(allUpdates) == 0 {
		fmt.Printf("No matching release found for image %s.\n", observeImage)
		return nil
	}
	if err = patchReleasesFilesMaybe(allUpdates, observePatch); err != nil {
		return err
	}
	return nil
}

func makeObserveReleaseFilters(args []string) []updates.ReleaseFilterFunc {
	filters := make([]updates.ReleaseFilterFunc, 0)
	if len(observeReleases) > 0 {
		filters = append(filters, updates.ReleaseFilter(observeReleases))
	}
	if len(args) == 1 {
		filters = append(filters, updates.EnvironmentReleaseFilter(args[0]))
	}
	if observeImage != "" {
		filters = append(filters, updates.ImageReleaseFilter(observeImage))
	}
	return filters
}

//func patchImageUpdates(releasesFile string, updates []kubecd.ImageUpdate) error {
//	var modYaml interface{}
//	data, err := ioutil.ReadFile(releasesFile)
//	if err != nil {
//		return errors.Wrapf(err, `error reading %q`, releasesFile)
//	}
//	err = yaml.Unmarshal(data, &modYaml)
//	if err != nil {
//		return errors.Wrapf(err, `error decoding yaml in %q`, releasesFile)
//	}
//	tmpReleases := modYaml.(map[interface{}]interface{})["releases"]
//	if tmpReleases == nil {
//		return fmt.Errorf(`%s: no "releases" found`, releasesFile)
//	}
//	releases, ok := tmpReleases.([]interface{})
//	if !ok {
//		return fmt.Errorf(`%s: "releases" not a list`, releasesFile)
//	}
//	for _, tmpRel := range releases {
//		modRel := tmpRel.(map[interface{}]interface{})
//		relName := modRel["name"].(string)
//		for _, update := range updates {
//			if relName != update.Release.Name {
//				continue
//			}
//			if _, found := modRel["values"]; !found {
//				modRel["values"] = make(map[interface{}]interface{})
//			}
//			foundVal := false
//			values := modRel["values"].([]interface{})
//			for _, tmpVal := range values {
//				modVal := tmpVal.(map[string]interface{})
//				if modVal["key"].(string) != update.TagValue {
//					continue
//				}
//				modVal["value"] = update.NewTag
//				foundVal = true
//				break
//			}
//			if !foundVal {
//				values = append(values, map[interface{}]interface{}{"key": update.TagValue, "value": update.NewTag})
//			}
//		}
//	}
//	if err = writeIndentedYamlToFile(releasesFile, modYaml); err != nil {
//		return err
//	}
//	return nil
//}

func patchReleasesFilesMaybe(imageUpdates []updates.ImageUpdate, patch bool) error {
	verb := "May"
	if patch {
		verb = "Will"
	}
	updatesPerFile := make(map[string][]updates.ImageUpdate)
	for _, update := range imageUpdates {
		fmt.Printf("%s update release %q image %q tag %s -> %s\n", verb, update.Release.Name, update.ImageRepo, update.OldTag, update.NewTag)
		file := update.Release.FromFile
		if _, found := updatesPerFile[file]; !found {
			updatesPerFile[file] = make([]updates.ImageUpdate, 0)
		}
		updatesPerFile[file] = append(updatesPerFile[file], update)
	}
	if patch {
		for file, fileUpdates := range updatesPerFile {
			fmt.Printf("Patching file: %s\n", file)
			if err := patchImageUpdatesYamlNode(file, fileUpdates); err != nil {
				return err
			}
		}
	}
	return nil
}

func patchImageUpdatesYamlNode(releasesFile string, imageUpdates []updates.ImageUpdate) error {
	var doc yaml.Node
	data, err := ioutil.ReadFile(releasesFile)
	if err != nil {
		return errors.Wrapf(err, `error reading %q`, releasesFile)
	}
	err = yaml.Unmarshal(data, &doc)
	if err != nil {
		return errors.Wrapf(err, `error decoding yaml in %q`, releasesFile)
	}
	releases := yamlNodeMapEntry(doc.Content[0], "releases")
	if releases.Kind != yaml.SequenceNode {
		return fmt.Errorf(`%s: "releases" is not a list`, releasesFile)
	}
	madeChanges := false
	for _, release := range releases.Content {
		name := yamlNodeMapEntry(release, "name")
		if name == nil || name.Kind != yaml.ScalarNode {
			continue
		}
		for _, update := range imageUpdates {
			if update.Release.Name != name.Value {
				continue
			}
			values := yamlNodeMapEntry(release, "values")
			if values == nil {
				continue
			}
			for _, chartValue := range values.Content {
				key := yamlNodeMapEntry(chartValue, "key")
				value := yamlNodeMapEntry(chartValue, "value")
				if key == nil || value == nil {
					continue
				}
				if key.Value == update.TagValue {
					value.Value = update.NewTag
					madeChanges = true
				}
			}
		}
	}
	if !madeChanges {
		return nil
	}
	return writeIndentedYamlToFile(releasesFile, &doc)
}

func yamlNodeMapEntry(node *yaml.Node, name string) *yaml.Node {
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Kind == yaml.ScalarNode && node.Content[i].Value == name {
				return node.Content[i+1]
			}
		}
	}
	return nil
}

func observeChartVersion(kcdConfig *model.KubeCDConfig) error {
	return NotYetImplementedError("observe --chart")
}

func imageOrChart(image, chart *string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if (*chart != "" && *image != "") || (*chart == "" && *image == "") {
			return errors.New("must specify --image or --chart")
		}
		return nil
	}
}
