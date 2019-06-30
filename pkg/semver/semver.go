package semver

import (
	"fmt"
	mmsemver "github.com/Masterminds/semver"
	"sort"
)

const (
	TrackPatchLevel = "PatchLevel"
	TrackMinorVersion = "MinorVersion"
	TrackMajorVersion = "MajorVersion"
	TrackNewest = "Newest"
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
		spec, err = mmsemver.NewConstraint("~" + current.String())
	case TrackMinorVersion:
		spec, err = mmsemver.NewConstraint("^" + current.String())
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
