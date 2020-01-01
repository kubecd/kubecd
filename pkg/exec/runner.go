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

package exec

import (
	"fmt"
	"os"
	osexec "os/exec"
	"strings"
)

type Runner interface {
	Run(string, ...string) ([]byte, error)
}

type RealRunner struct{}

func (r RealRunner) Run(cmd string, args ...string) ([]byte, error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", cmd, strings.Join(args, " "))
	return osexec.Command(cmd, args...).Output()
}
