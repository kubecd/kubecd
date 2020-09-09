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
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

var mockedExitStatus = 0
var mockedStdout string

func TestChartValue_UnmarshalJSON(t *testing.T) {
	var cv1, cv2 ChartValue
	assert.NoError(t, json.Unmarshal([]byte(`{"key": "foo", "value": "bar"}`), &cv1))
	assert.Equal(t, "bar", cv1.Value)
	assert.NoError(t, json.Unmarshal([]byte(`{"key": "foo", "value": 42}`), &cv2))
	assert.Equal(t, "42", cv2.Value)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	es := strconv.Itoa(mockedExitStatus)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		"EXIT_STATUS=" + es}
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// println("Mocked stdout:", os.Getenv("STDOUT"))
	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

func TestHelmVersion_GetMajorVersion(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	mockedExitStatus = 0
	mockedStdout = "v3.3.0+g8a4aeec"

	assert.Equal(t, 3, HelmVersion{Path: "helm"}.GetMajorVersion())

	mockedExitStatus = 0
	mockedStdout = "v2.3.0+g8a4aeec"

	assert.Equal(t, 2, HelmVersion{Path: "helm"}.GetMajorVersion())

	mockedExitStatus = 0
	mockedStdout = "1.3.0+asdasd"

	assert.Equal(t, 1, HelmVersion{Path: "helm"}.GetMajorVersion())

	mockedExitStatus = 0
	mockedStdout = "Client: v2.9.1+g20adb27"
	assert.Equal(t, 2, HelmVersion{Path: "helm"}.GetMajorVersion())
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
