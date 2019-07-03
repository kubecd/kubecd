package model

import (
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRelease_UnmarshalJSON(t *testing.T) {
	yamlData := []byte(`
name: release1
trigger:
  image:
    track: Newest
`)
	var release Release
	require.NoError(t, yaml.Unmarshal(yamlData, &release))
	assert.Nil(t, release.Trigger)
	assert.Len(t, release.Triggers, 1)
}
