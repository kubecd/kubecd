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
