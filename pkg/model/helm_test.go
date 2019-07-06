package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestChartValue_UnmarshalJSON(t *testing.T) {
	var cv1, cv2 ChartValue
	assert.NoError(t, json.Unmarshal([]byte(`{"key": "foo", "value": "bar"}`), &cv1))
	assert.Equal(t, "bar", cv1.Value)
	assert.NoError(t, json.Unmarshal([]byte(`{"key": "foo", "value": 42}`), &cv2))
	assert.Equal(t, "42", cv2.Value)
}

func TestFlexString_UnmarshalJSON(t *testing.T) {
	type testCase struct {
		data     string
		expected string
	}
	for i, tc := range []testCase{
		{`"42"`, "42"},
		{`43`, "43"},
		{`true`, "true"},
		{`"true"`, "true"},
		{`false`, "false"},
		{`"false"`, "false"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var fs FlexString
			assert.NoError(t, json.Unmarshal([]byte(tc.data), &fs))
			assert.Equal(t, tc.expected, string(fs))
		})

	}
}
