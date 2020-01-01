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

package updates

import (
	"github.com/kubecd/kubecd/pkg/image"
	"github.com/kubecd/kubecd/pkg/model"
	"github.com/kubecd/kubecd/pkg/semver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImageReleaseIndex(t *testing.T) {
	env := &model.Environment{Name: "test"}
	releases := []*model.Release{
		{
			Name: "release1",
			Values: []model.ChartValue{
				{Key: model.DefaultRepoValue, Value: "test-image"},
				{Key: model.DefaultTagValue, Value: "v1.0"},
			},
			Triggers:    []model.ReleaseUpdateTrigger{{Image: &model.ImageTrigger{Track: semver.TrackPatchLevel}}},
			Environment: env,
		},
		{
			Name: "release2",
			Values: []model.ChartValue{
				{Key: model.DefaultRepoValue, Value: "test-image"},
				{Key: model.DefaultTagValue, Value: "v1.1"},
			},
			Triggers:    []model.ReleaseUpdateTrigger{{Image: &model.ImageTrigger{Track: semver.TrackPatchLevel}}},
			Environment: env,
		},
		{
			Name: "release3",
			Values: []model.ChartValue{
				{Key: model.DefaultRepoValue, Value: "test-image2"},
				{Key: model.DefaultTagValue, Value: "v1.0"},
			},
			Triggers:    []model.ReleaseUpdateTrigger{{Image: &model.ImageTrigger{Track: semver.TrackPatchLevel}}},
			Environment: env,
		},
	}
	env.Releases = releases
	kcdConfig := &model.KubeCDConfig{Environments: []*model.Environment{env}}
	index, err := ImageReleaseIndex(kcdConfig)
	assert.NoError(t, err)
	assert.Len(t, index, 2)
	assert.Len(t, index[image.DefaultDockerRegistry+"/test-image"], 2)
	assert.Len(t, index[image.DefaultDockerRegistry+"/test-image2"], 1)
	assert.Equal(t, "release1", index[image.DefaultDockerRegistry+"/test-image"][0].Name)
	assert.Equal(t, "release2", index[image.DefaultDockerRegistry+"/test-image"][1].Name)
	assert.Equal(t, "release3", index[image.DefaultDockerRegistry+"/test-image2"][0].Name)
}
