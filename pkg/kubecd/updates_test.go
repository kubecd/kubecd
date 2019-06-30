package kubecd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseDockerTimetsamp(t *testing.T) {
	_, err := parseDockerTimestamp("bad")
	assert.Error(t, err)
	ts, err := parseDockerTimestamp("2019-06-29T08:47:03.62062787Z02:00")
	assert.NoError(t, err)
	assert.Equal(t, int64(1561790823), ts)
}
