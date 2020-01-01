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
		{`0.1`, "0.1"},
		{`"string"`, "string"},
		{`0.100`, "0.1"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var fs FlexString
			assert.NoError(t, json.Unmarshal([]byte(tc.data), &fs))
			assert.Equal(t, tc.expected, string(fs))
		})

	}
}
