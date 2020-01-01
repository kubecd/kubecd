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

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	. "github.com/bitfield/script"
)

var remotesRegexp = regexp.MustCompile(`([a-z0-9A-Z_.-]+)\t.*github.com[/:]zedge/kubecd.*\(fetch\)`)
var lsRemoteRegexp = regexp.MustCompile(`.*/v`)

type bumpMode string

const (
	bumpMajor   bumpMode = "major"
	bumpMinor   bumpMode = "minor"
	bumpPatch   bumpMode = "patch"
	defaultBump          = bumpMinor
)

func (m bumpMode) Bump(ver semver.Version) string {
	var next semver.Version
	switch m {
	case bumpMajor:
		next = ver.IncMajor()
	case bumpMinor:
		next = ver.IncMinor()
	case bumpPatch:
		next = ver.IncPatch()
	default:
		panic(fmt.Sprintf("invalid bumpMode: %q", m))
	}
	return (&next).String()
}

func main() {
	remoteName := getOriginRemoteName()
	latestVersion := getMostRecentRemoteTag(remoteName)

	bump := defaultBump
	if len(os.Args) == 2 {
		bump = bumpMode(os.Args[1])
	}
	fmt.Printf("v%s", bump.Bump(*latestVersion))
}

func getMostRecentRemoteTag(remoteName string) *semver.Version {
	latestVersion, _ := semver.NewVersion("0.0.0")
	Exec("git ls-remote "+strings.TrimSpace(remoteName)+" refs/tags/v*").ReplaceRegexp(lsRemoteRegexp, "").EachLine(func(s string, _ *strings.Builder) {
		if ver, _ := semver.NewVersion(s); ver != nil && ver.GreaterThan(latestVersion) {
			latestVersion = ver
		}
	})
	return latestVersion
}

func getOriginRemoteName() string {
	remoteName, _ := Exec("git remote -v").ReplaceRegexp(remotesRegexp, `$1`).First(1).String()
	return strings.TrimSpace(remoteName)
}
