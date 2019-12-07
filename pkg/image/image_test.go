package image

import (
	"github.com/kubecd/kubecd/pkg/exec"
	"github.com/kubecd/kubecd/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestGetNewestMatchingTag(t *testing.T) {
	type testCase struct {
		current    string
		candidates []string
		track      string
		expected   string
	}
	for i, tc := range []testCase{
		{"1.0", []string{"0.9.0", "1.0.1", "1.1", "2.0"}, semver.TrackPatchLevel, "1.0.1"},
		{"1.0", []string{"0.9.0", "1.0.1", "1.1", "2.0"}, semver.TrackMinorVersion, "1.1"},
		{"1.0", []string{"0.9.0", "1.0.1", "1.1", "2.0"}, semver.TrackMajorVersion, "2.0"},
		{"1.0", []string{"0.9.0", "v1.0.1", "1.1", "2.0"}, semver.TrackPatchLevel, "v1.0.1"},
		{"1.0", []string{"0.9.0", "1.0.1", "v1.1", "2.0"}, semver.TrackMinorVersion, "v1.1"},
		{"1.0", []string{"0.9.0", "1.0.1", "1.1", "v2.0"}, semver.TrackMajorVersion, "v2.0"},
		{"foo", []string{"c", "b", "a"}, semver.TrackNewest, "a"},
		{"foo", []string{"a", "b", "c"}, semver.TrackNewest, "c"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			current := TimestampedTag{Tag: tc.current}
			candidates := make([]TimestampedTag, len(tc.candidates))
			for j, tag := range tc.candidates {
				candidates[j] = TimestampedTag{Tag: tag, Timestamp: int64(j)}
			}
			expected := TimestampedTag{Tag: tc.expected}
			assert.Equal(t, expected.Tag, GetNewestMatchingTag(current, candidates, tc.track).Tag)
		})
	}
}

func TestParseDockerTimestamp(t *testing.T) {
	_, err := ParseDockerTimestamp("bad")
	assert.Error(t, err)
	ts, err := ParseDockerTimestamp("2015-02-13T10:04:55.62062787Z")
	assert.NoError(t, err)
	assert.Equal(t, int64(1423821895), ts)
}

func TestGetTagsForGcrImage(t *testing.T) {
	fixture, err := ioutil.ReadFile("testdata/gcr-airflow-tags.json")
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
