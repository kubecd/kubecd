package semver

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsSemver(t *testing.T) {
	assert.True(t, IsSemver("v1.0"))
	assert.False(t, IsSemver("deadbeef"))
}

func TestBestUpgrade(t *testing.T) {
	type testCase struct {
		current    string
		track      string
		candidates []string
		best       string
		error      error
	}
	for _, tc := range []testCase{
		{"1.0.0", TrackMinorVersion, []string{"0.9.0", "2.0.0", "1.0.1", "1.2.0"}, "1.2.0", nil},
		{"1.0.0", TrackMajorVersion, []string{"0.9.0", "2.0.0", "1.0.1", "1.2.0"}, "2.0.0", nil},
		{"1.0.0", TrackPatchLevel, []string{"0.9.0", "2.0.0", "1.0.1", "1.2.0"}, "1.0.1", nil},
		{"1.0.3", TrackPatchLevel, []string{"0.9.0", "2.0.0", "1.0.1", "1.2.0"}, "", fmt.Errorf(`found no versions >1.0.3`)},
	} {
		current, _ := Parse(tc.current)
		expected, _ := Parse(tc.best)
		candidates := make([]*semver.Version, len(tc.candidates))
		for i, c := range tc.candidates {
			candidates[i], _ = Parse(c)
		}
		best, err := BestUpgrade(current, candidates, tc.track)
		if tc.error != nil {
			assert.Error(t, err)
			assert.Equal(t, tc.error, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, expected, best)
		}
	}
	v1, _ := Parse("v1.0")
	_, err := BestUpgrade(v1, []*semver.Version{v1}, "UnknownTrack")
	assert.Error(t, err)
	assert.Equal(t, `unknown "track": "UnknownTrack"`, err.Error())
}

func TestIsWantedUpgrade(t *testing.T) {
	type testCase struct {
		current   string
		candidate string
		track     string
		isWanted  bool
	}
	for i, tc := range []testCase{
		{"1.0", "0.9", TrackMajorVersion, false},
		{"1.0", "1.1", TrackMinorVersion, true},
		{"1.0", "1.0", TrackMinorVersion, false},
		{"1.0", "1.0", "UnknownTrack", false},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			current, _ := Parse(tc.current)
			candidate, _ := Parse(tc.candidate)
			wanted := IsWantedUpgrade(current, candidate, tc.track)
			assert.Equal(t, tc.isWanted, wanted)
		})
	}
}
