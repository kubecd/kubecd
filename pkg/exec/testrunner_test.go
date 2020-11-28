/*
 * Copyright 2018-2019 Zedge, Inc.
 * Copyright 2019-2020 Stig Sæther Nordahl Bakken
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

package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ Runner = TestRunner{}

func TestTestRunner(t *testing.T) {
	expected := "testing the test runner"
	runner := TestRunner{Output: []byte(expected)}
	output, err := runner.Run("command")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(output))
	runner = TestRunner{ExitCode: 42}
	_, err = runner.Run("command")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "exit status 42")
	runner = TestRunner{ExpectedCommand: []string{"foo", "bar"}}
	output, err = runner.Run("foo", "bar", "gazonk")
	assert.Error(t, err)
	assert.Equal(t, "exit status 127", err.Error())
	assert.Equal(t, "expected argv [foo bar], got [foo bar gazonk]", string(output))
}

func TestHelperProcess(t *testing.T) {
	InsideHelperProcess()
}
