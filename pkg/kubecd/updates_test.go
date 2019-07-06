package kubecd

import (
	"github.com/stretchr/testify/assert"
	"github.com/zedge/kubecd/pkg/image"
	"github.com/zedge/kubecd/pkg/model"
	"github.com/zedge/kubecd/pkg/semver"
	"testing"
)

//func TestReleaseWantsImageUpdate(t *testing.T) {
//	testImage1 := "test-image1"
//	testImage2 := "test-image2"
//	currentTag1 := "v1.0"
//	newTag1 := "v1.0.1"
//	release := &model.Release{
//		Triggers: []model.ReleaseUpdateTrigger{
//			{Image: &model.ImageTrigger{Track: semver.TrackMinorVersion}},
//		},
//		Values: []model.ChartValue{
//			{Key: model.DefaultRepoValue, Value: testImage1},
//			{Key: model.DefaultTagValue, Value: currentTag1},
//		},
//	}
//	newImageRef1 := &image.DockerImageRef{Image: testImage1, Tag: newTag1}
//	updates1, err := ReleaseWantsImageUpdate(release, newImageRef1)
//	assert.NoError(t, err)
//	require.Len(t, updates1, 1)
//	updates1[0].Release = nil
//	assert.Equal(t, ImageUpdate{NewTag: newTag1, TagValue: model.DefaultTagValue, ImageRepo: testImage1, Reason: `with track="MinorVersion", "v1.0.1" > "v1.0"`}, updates1[0])
//	newImageRef2 := &image.DockerImageRef{Image: testImage2, Tag: newTag1}
//	updates2, err := ReleaseWantsImageUpdate(release, newImageRef2)
//	assert.NoError(t, err)
//	assert.Len(t, updates2, 0)
//}

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
	assert.Len(t, index[image.DefaultDockerRegistry + "/test-image"], 2)
	assert.Len(t, index[image.DefaultDockerRegistry + "/test-image2"], 1)
	assert.Equal(t, "release1", index[image.DefaultDockerRegistry + "/test-image"][0].Name)
	assert.Equal(t, "release2", index[image.DefaultDockerRegistry + "/test-image"][1].Name)
	assert.Equal(t, "release3", index[image.DefaultDockerRegistry + "/test-image2"][0].Name)
}
