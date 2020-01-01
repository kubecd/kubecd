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

package semver

import (
	"fmt"
	mmsemver "github.com/Masterminds/semver"
	"sort"
)

const (
	TrackPatchLevel   = "PatchLevel"
	TrackMinorVersion = "MinorVersion"
	TrackMajorVersion = "MajorVersion"
	TrackNewest       = "Newest"
)

func IsSemver(version string) bool {
	_, err := mmsemver.NewVersion(Normalize(version))
	return err == nil
}

func Normalize(version string) string {
	if len(version) > 0 && version[0] == 'v' {
		return version[1:]
	}
	return version
}

func Parse(version string) (*mmsemver.Version, error) {
	return mmsemver.NewVersion(Normalize(version))
}

func BestUpgrade(current *mmsemver.Version, candidates []*mmsemver.Version, track string) (*mmsemver.Version, error) {
	var spec *mmsemver.Constraints
	var err error
	switch track {
	case TrackPatchLevel:
		nextMinor := (*current).IncMinor()
		spec, err = mmsemver.NewConstraint(">" + current.String() + ", <" + nextMinor.String())
	case TrackMinorVersion:
		nextMajor := (*current).IncMajor()
		spec, err = mmsemver.NewConstraint(">" + current.String() + ", <" + nextMajor.String())
	case TrackMajorVersion:
		spec, err = mmsemver.NewConstraint(">" + current.String())
	default:
		return nil, fmt.Errorf(`unknown "track": %q`, track)
	}
	if err != nil {
		return nil, err
	}
	filtered := make([]*mmsemver.Version, 0)
	for _, c := range candidates {
		if spec.Check(c) {
			filtered = append(filtered, c)
		}
	}
	sort.Sort(mmsemver.Collection(filtered))
	if len(filtered) > 0 {
		return filtered[len(filtered)-1], nil
	}
	return nil, fmt.Errorf(`found no versions >%s`, current.String())
}

func IsWantedUpgrade(current, candidate *mmsemver.Version, track string) bool {
	_, err := BestUpgrade(current, []*mmsemver.Version{candidate}, track)
	return err == nil
}
