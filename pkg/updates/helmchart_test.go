package updates

import (
	"github.com/stretchr/testify/assert"
	"github.com/zedge/kubecd/pkg/image"
	"github.com/zedge/kubecd/pkg/model"
	"github.com/zedge/kubecd/pkg/semver"
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
