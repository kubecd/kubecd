package kubecd

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zedge/kubecd/pkg/exec"
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
			"gcloud", "container", "images", "list-tags", imageRef.ImageStringNoTag(),
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

// TestHelperProcess is required boilerplate (one per package) for using exec.TestRunner
func TestHelperProcess(t *testing.T) {
	exec.InsideHelperProcess()
}
