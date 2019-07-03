package kubecd

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zedge/kubecd/pkg/exec"
	"github.com/zedge/kubecd/pkg/model"
	"github.com/zedge/kubecd/pkg/semver"
	"io/ioutil"
	"testing"
)

func TestParseDockerTimestamp(t *testing.T) {
	_, err := parseDockerTimestamp("bad")
	assert.Error(t, err)
	ts, err := parseDockerTimestamp("2015-02-13T10:04:55.62062787Z")
	assert.NoError(t, err)
	assert.Equal(t, int64(1423821895), ts)
}

func TestGetTagsForGcrImage(t *testing.T) {
	fixture, err := ioutil.ReadFile("../../kubecd/tests/testdata/gcr-airflow-tags.json")
	require.NoError(t, err)
	imageRef := &DockerImageRef{Registry: "gcr.io", Image: "my-project/some-image"}
	runner = exec.TestRunner{
		Output: fixture,
		ExpectedCommand: []string{
			"gcloud", "container", "images", "list-tags", imageRef.WithoutTag(),
			"--project", "my-project", "--format", "json",
		},
	}
	timestampedTags, err := GetTagsForGcrImage(imageRef)
	assert.NoError(t, err)
	assert.Len(t, timestampedTags, 2)
	assert.Equal(t, int64(1516372388), timestampedTags[0].Timestamp)
	assert.Equal(t, int64(1514653063), timestampedTags[1].Timestamp)
	assert.Equal(t, "1.8.2", timestampedTags[0].Tag)
	assert.Equal(t, "1.8", timestampedTags[1].Tag)
}

func TestReleaseWantsImageUpdate(t *testing.T) {
	testImage1 := "test-image1"
	testImage2 := "test-image2"
	currentTag1 := "v1.0"
	newTag1 := "v1.0.1"
	release := &model.Release{
		Triggers: []model.ReleaseUpdateTrigger{
			{Image: &model.ImageTrigger{Track: semver.TrackMinorVersion}},
		},
		Values: []model.ChartValue{
			{Key: "image.repository", Value: testImage1},
			{Key: "image.tag", Value: currentTag1},
		},
	}
	newImageRef1 := &DockerImageRef{Image: testImage1, Tag: newTag1}
	updates1, err := ReleaseWantsImageUpdate(release, newImageRef1)
	assert.NoError(t, err)
	assert.Len(t, updates1, 1)
	updates1[0].Release = nil
	assert.Equal(t, ImageUpdate{NewTag: newTag1, TagValue: model.DefaultTagValue, ImageRepo: testImage1, Reason: `with track="MinorVersion", "v1.0.1" > "v1.0"`}, updates1[0])
	newImageRef2 := &DockerImageRef{Image: testImage2, Tag: newTag1}
	updates2, err := ReleaseWantsImageUpdate(release, newImageRef2)
	assert.NoError(t, err)
	assert.Len(t, updates2, 0)
}

// TestHelperProcess is required boilerplate (one per package) for using exec.TestRunner
func TestHelperProcess(t *testing.T) {
	exec.InsideHelperProcess()
}

func TestNewDockerImageRef(t *testing.T) {
	type testCase struct {
		expected DockerImageRef
		image    string
	}
	for _, tc := range []testCase{
		{expected: DockerImageRef{Registry: "eu.gcr.io", Image: "kubecd-demo/prod-demo-app", Tag: "v1.1"}, image: "eu.gcr.io/kubecd-demo/prod-demo-app:v1.1"},
		{expected: DockerImageRef{Registry: "eu.gcr.io", Image: "kubecd-demo/prod-demo-app", Tag: ""}, image: "eu.gcr.io/kubecd-demo/prod-demo-app"},
		{expected: DockerImageRef{Registry: DefaultDockerRegistry, Image: "zedge/kubecd", Tag: "latest"}, image: "zedge/kubecd:latest"},
	} {
		ref := NewDockerImageRef(tc.image)
		assert.Equal(t, tc.expected, *ref)
	}
}
