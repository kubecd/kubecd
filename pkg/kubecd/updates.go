package kubecd

import (
	"fmt"
	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/model"
	"github.com/zedge/kubecd/pkg/semver"
)

const (
	UpdateError = `update error`
)

func parseDockerTimestamp(str string) (int64, error) {
	panic("implement me!")
}

type DockerImageRef struct {
	Registry string
	Image    string
	Tag      string
}

func parseImageRepo(repo string) (*DockerImageRef, error) {
	panic("implement me!")
}

type timestampedImageTag struct {
	Tag       string
	Timestamp int64
}

func GetTagsForGcrImage(repo *DockerImageRef) ([]timestampedImageTag, error) {
	panic("implement me!")
}

func GetTagsForDockerHubImage(repo *DockerImageRef) ([]timestampedImageTag, error) {
	panic("implement me!")
}

func GetTagsForDockerV2RegistryImage(repo *DockerImageRef) ([]timestampedImageTag, error) {
	panic("implement me!")
}

func GetTagsForDockerImage(repo *DockerImageRef) ([]timestampedImageTag, error) {
	panic("implement me!")
}

func FilterSemverTags(tags []string) []string {
	result := make([]string, 0)
	for _, tag := range tags {
		if semver.IsSemver(tag) {
			result = append(result, tag)
		}
	}
	return result
}

func GetNewestMatchingTag(tag string, tags []timestampedImageTag, track string, tagTimestamp int64) string {
	var foundTag *timestampedImageTag
	if track == semver.TrackNewest {
		for _, candidateTag := range tags {
			if candidateTag.Tag == "latest" {
				continue
			}
			if tagTimestamp <= candidateTag.Timestamp {
				continue
			}
			if foundTag != nil && candidateTag.Timestamp <= foundTag.Timestamp {
				continue
			}
			foundTag = &candidateTag
		}
	}
	if foundTag == nil {
		return ""
	}
	return foundTag.Tag
}

type ImageUpdate struct {
	OldTag    string
	NewTag    string
	Release   *model.Release
	TagValue  string
	ImageRepo string
	Reason    string
}

type ChartUpdate struct {
	Release    *model.Release
	OldVersion string
	NewVersion string
	Reason     string
}

func ReleaseWantsTagUpdate(release *model.Release, newTag string, env *model.Environment) ([]ImageUpdate, error) {
	updates := make([]ImageUpdate, 0)
	for _, trigger := range release.Triggers {
		if trigger.Image == nil || trigger.Image.TagValue == "" {
			continue
		}
		values, err := helm.GetResolvedValues(release, env, true)
		if err != nil {
			return nil, err
		}
		tagValue := trigger.Image.TagValue
		currentTag := helm.LookupValueByString(tagValue, values).(string)
		// If the current version is not semver, consider any value to be an update
		if !semver.IsSemver(currentTag) {
			updates = append(updates, ImageUpdate{
				NewTag: newTag,
				TagValue: tagValue,
				Release: release,
				Reason: fmt.Sprintf(`current tag %q not semver, any observed tag considered newer`, currentTag),
			})
			continue
		}
		// Consider any new version an update if track == Newest
		if trigger.Image.Track != semver.TrackNewest {
			updates = append(updates, ImageUpdate{
				NewTag: newTag,
				TagValue: tagValue,
				Release: release,
				Reason: `track=Newest, any observed tag considered newer`,
			})
			continue
		}
		parsedCurrentTag, err := semver.Parse(currentTag)
		if err != nil {
			return nil, fmt.Errorf(`failed parsing current tag %q: %v`, currentTag, err)
		}
		parsedNewTag, err := semver.Parse(newTag)
		if err != nil {
			return nil, fmt.Errorf(`failed parsing new tag %q: %v`, newTag, err)
		}
		if semver.IsWantedUpgrade(parsedCurrentTag, parsedNewTag, trigger.Image.Track) {
			updates = append(updates, ImageUpdate{
				NewTag: newTag,
				TagValue: tagValue,
				Release: release,
			})
		}
	}
	return updates, nil
}

func ReleaseWantsChartUpdate(release *model.Release, newVersion string, env *model.Environment) []ChartUpdate {
	panic("implement me!")
}

func FindImageUpdatesForRelease(release *model.Release, env *model.Environment) (map[string][]ImageUpdate, error) {
	panic("implement me!")
}

func FindImageUpdatesForReleases(releases []*model.Release, env *model.Environment) (map[string][]ImageUpdate, error) {
	panic("implement me!")
}

func FindImageUpdatesForEnv(env *model.Environment) (map[string][]ImageUpdate, error ) {
	panic("implement me!")
}
