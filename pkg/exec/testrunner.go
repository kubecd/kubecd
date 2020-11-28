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
	"encoding/json"
	"fmt"
	"os"
	osexec "os/exec"
	"reflect"
	"strconv"
)

type TestRunner struct {
	ExpectedCommand []string
	Output          []byte
	ExtraEnv        map[string]string
	ExitCode        int
}

func (r TestRunner) Run(command string, args ...string) ([]byte, error) {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := osexec.Command(os.Args[0], cs...)
	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"GO_HELPER_MOCK_STDOUT=" + string(r.Output), // may need to base64 encode here?
		fmt.Sprintf("GO_HELPER_MOCK_EXIT_CODE=%d", r.ExitCode),
	}
	if r.ExtraEnv != nil {
		for k, v := range r.ExtraEnv {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
	if r.ExpectedCommand != nil {
		jsonArr, err := json.Marshal(r.ExpectedCommand)
		if err != nil {
			return nil, err
		}
		cmd.Env = append(cmd.Env, "GO_HELPER_EXPECTED_COMMAND_JSON="+string(jsonArr))
	}
	out, err := cmd.CombinedOutput()
	return out, err
}

func InsideHelperProcess() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	exitCode, err := strconv.Atoi(os.Getenv("GO_HELPER_MOCK_EXIT_CODE"))
	if err != nil {
		panic(err)
	}
	if expCmdEnv, found := os.LookupEnv("GO_HELPER_EXPECTED_COMMAND_JSON"); found {
		var expectedCommand []string
		actualCommand := os.Args[3:]
		err = json.Unmarshal([]byte(expCmdEnv), &expectedCommand)
		if err != nil {
			panic(err)
		}
		if len(expectedCommand) > 0 {
			if !reflect.DeepEqual(actualCommand, expectedCommand) {
				fmt.Printf("expected argv %v, got %v", expectedCommand, actualCommand)
				os.Exit(127)
			}
		}
	}
	fmt.Print(os.Getenv("GO_HELPER_MOCK_STDOUT"))
	os.Exit(exitCode)
}
