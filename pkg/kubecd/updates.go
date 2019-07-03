package kubecd

import (
	"encoding/json"
	"fmt"
	registry2 "github.com/heroku/docker-registry-client/registry"
	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/model"
	"github.com/zedge/kubecd/pkg/semver"
	"strings"
	"time"
)

const (
	DefaultDockerRegistry = "registry-1.docker.io"
	GCRRegistrySuffix     = "gcr.io"
)

func parseDockerTimestamp(str string) (int64, error) {
	timestamp, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		return int64(0), err
	}
	return timestamp.Unix(), nil
}

type DockerImageRef struct {
	Registry string
	Image    string
	Tag      string
}

func NewDockerImageRef(repo string) *DockerImageRef {
	return parseImageRepo(repo)
}

func (r *DockerImageRef) RegistryURL() string {
	return "https://" + r.Registry
}

func (r *DockerImageRef) WithTag() string {
	return r.WithoutTag() + ":" + r.Tag
}

func (r *DockerImageRef) WithoutTag() string {
	tmp := ""
	if r.Registry != "" {
		tmp += r.Registry + "/"
	}
	return tmp + r.Image
}

func parseImageRepo(repo string) *DockerImageRef {
	result := &DockerImageRef{}
	tmp := strings.Split(repo, "/")
	if len(tmp) > 1 && strings.IndexByte(tmp[0], '.') != -1 {
		result.Registry = tmp[0]
		tmp = tmp[1:]
	}
	lastElem := tmp[len(tmp)-1]
	colonIndex := strings.IndexByte(lastElem, ':')
	if colonIndex != -1 {
		result.Tag = lastElem[colonIndex+1:]
		tmp[len(tmp)-1] = lastElem[0:colonIndex]
	}
	result.Image = strings.Join(tmp, "/")
	if result.Registry == "" {
		result.Registry = DefaultDockerRegistry
	}
	return result
}

type timestampedImageTag struct {
	Tag       string
	Timestamp int64
}

func GetTagsForGcrImage(repo *DockerImageRef) ([]timestampedImageTag, error) {
	fullRepo := repo.WithoutTag()
	tmp := strings.Split(repo.Image, "/")
	gcpProject := tmp[0]
	output, err := runner.Run("gcloud", "container", "images", "list-tags", fullRepo, "--project", gcpProject, "--format", "json")
	if err != nil {
		return nil, fmt.Errorf(`failed listing tags for GCR image %q: %v`, fullRepo, err)
	}
	result := make([]timestampedImageTag, 0)
	type gcrTimestamp struct {
		Year   int `json:"year"`
		Month  int `json:"month"`
		Day    int `json:"day"`
		Hour   int `json:"hour"`
		Minute int `json:"minute"`
		Second int `json:"second"`
	}
	type gcrImageTag struct {
		Digest    string       `json:"digest"`
		Tags      []string     `json:"tags"`
		Timestamp gcrTimestamp `json:"timestamp"`
	}
	var imageTagList []gcrImageTag
	err = json.Unmarshal(output, &imageTagList)
	if err != nil {
		return nil, fmt.Errorf(`failed decoding output when getting tags for GCR image %q: %v`, fullRepo, err)
	}
	for _, imgTag := range imageTagList {
		ts := imgTag.Timestamp
		timestamp := time.Date(ts.Year, time.Month(ts.Month+1), ts.Day, ts.Hour, ts.Minute, ts.Second, 0, time.UTC)
		for _, tag := range imgTag.Tags {
			result = append(result, timestampedImageTag{Tag: tag, Timestamp: timestamp.Unix()})
		}
	}
	return result, nil
}

func GetTagsForDockerHubImage(repo *DockerImageRef) ([]timestampedImageTag, error) {
	return GetTagsForDockerV2RegistryImage(repo)
}

func GetTagsForDockerV2RegistryImage(repo *DockerImageRef) ([]timestampedImageTag, error) {
	registry, err := registry2.New(repo.RegistryURL(), "", "")
	result := make([]timestampedImageTag, 0)
	if err != nil {
		return nil, fmt.Errorf(`could not access Docker registry %q: %v`, repo.Registry, err)
	}
	tags, err := registry.Tags(repo.Image)
	if err != nil {
		return nil, fmt.Errorf(`could not list tags for %s: %v`, repo.WithoutTag(), err)
	}
	for _, tag := range tags {
		manifest, err := registry.Manifest(repo.Image, tag)
		if err != nil {
			return nil, fmt.Errorf(`could not get manifest for %s:%s: %v`, repo.WithoutTag(), tag, err)
		}
		type v1Compat struct {
			Created string `json:"created"`
		}
		if len(manifest.History) > 0 {
			var v1Compat v1Compat
			if err = json.Unmarshal([]byte(manifest.History[0].V1Compatibility), &v1Compat); err != nil {
				return nil, fmt.Errorf(`could not decode tag timestamp for %s:%s: %v`, repo.WithoutTag(), tag, err)
			}
			//timestamp, err := time.Parse("2006-01-02T15:04:05.999999999Z", v1Compat.Created)
			timestamp, err := parseDockerTimestamp(v1Compat.Created)
			if err != nil {
				return nil, fmt.Errorf(`could not parse timestamp for %s:%s: %v`, repo.WithoutTag(), tag, err)
			}
			result = append(result, timestampedImageTag{Tag: tag, Timestamp: timestamp})
		}
	}
	return result, nil
}

func GetTagsForDockerImage(repo string) ([]timestampedImageTag, error) {
	imgRef := parseImageRepo(repo)
	if strings.HasSuffix(imgRef.Registry, GCRRegistrySuffix) {
		return GetTagsForGcrImage(imgRef)
	}
	if imgRef.Registry == DefaultDockerRegistry {
		return GetTagsForDockerHubImage(imgRef)
	}
	return GetTagsForDockerV2RegistryImage(imgRef)
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

func ReleaseWantsImageUpdate(release *model.Release, newImage *DockerImageRef) ([]ImageUpdate, error) {
	updates := make([]ImageUpdate, 0)
	for _, trigger := range release.Triggers {
		if trigger.Image == nil {
			//fmt.Printf("no image trigger: release %q", release.Name)
			continue
		}
		values, err := helm.GetResolvedValues(release)
		if err != nil {
			return nil, err
		}
		imageRepo := helm.GetImageRepoFromImageTrigger(trigger.Image, values)
		if imageRepo != newImage.WithoutTag() {
			//fmt.Printf("imageRepo != newImage: %q != %q", imageRepo, newImage.WithoutTag())
			continue
		}
		tagValue := trigger.Image.TagValueString()
		currentTag := helm.LookupValueByString(tagValue, values).(*string)
		if currentTag == nil {
			continue
		}
		// If the current version is not semver, or track is "Newest", blindly treat any value as an update
		if !semver.IsSemver(*currentTag) || trigger.Image.Track == semver.TrackNewest {
			reason := `track=Newest, any observed tag considered newer`
			if trigger.Image.Track != semver.TrackNewest {
				reason = fmt.Sprintf(`current tag %q not semver, any observed tag considered newer`, *currentTag)
			}
			updates = append(updates, ImageUpdate{
				ImageRepo: imageRepo,
				NewTag:    newImage.Tag,
				TagValue:  tagValue,
				Release:   release,
				Reason:    reason,
			})
			continue
		}
		parsedCurrentTag, err := semver.Parse(*currentTag)
		if err != nil {
			return nil, fmt.Errorf(`release %q: failed parsing current tag %q: %v`, release.Name, *currentTag, err)
		}
		parsedNewTag, err := semver.Parse(newImage.Tag)
		if err != nil {
			return nil, fmt.Errorf(`release %q: failed parsing new tag %q: %v`, release.Name, newImage.Tag, err)
		}
		if semver.IsWantedUpgrade(parsedCurrentTag, parsedNewTag, trigger.Image.Track) {
			updates = append(updates, ImageUpdate{
				ImageRepo: imageRepo,
				NewTag:    newImage.Tag,
				TagValue:  tagValue,
				Release:   release,
				Reason:    fmt.Sprintf(`with track=%q, %q > %q`, trigger.Image.Track, newImage.Tag, *currentTag),
			})
		}
	}
	return updates, nil
}

//func ReleaseWantsTagUpdate(release *model.Release, newTag string) ([]ImageUpdate, error) {
//	updates := make([]ImageUpdate, 0)
//	for _, trigger := range release.Triggers {
//		if trigger.Image == nil {
//			continue
//		}
//		values, err := helm.GetResolvedValues(release)
//		if err != nil {
//			return nil, err
//		}
//		tagValue := trigger.Image.TagValueString()
//		currentTag := helm.LookupValueByString(tagValue, values).(*string)
//		imageRepo := helm.GetImageRepoFromImageTrigger(trigger.Image, values)
//		// If the current version is not semver, consider any value to be an update
//		if !semver.IsSemver(*currentTag) {
//			updates = append(updates, ImageUpdate{
//				ImageRepo: imageRepo,
//				NewTag:    newTag,
//				TagValue:  tagValue,
//				Release:   release,
//				Reason:    fmt.Sprintf(`current tag %q not semver, any observed tag considered newer`, *currentTag),
//			})
//			continue
//		}
//		// Consider any new version an update if track == Newest
//		if trigger.Image.Track != semver.TrackNewest {
//			updates = append(updates, ImageUpdate{
//				ImageRepo: imageRepo,
//				NewTag:    newTag,
//				TagValue:  tagValue,
//				Release:   release,
//				Reason:    `track=Newest, any observed tag considered newer`,
//			})
//			continue
//		}
//		parsedCurrentTag, err := semver.Parse(*currentTag)
//		if err != nil {
//			return nil, fmt.Errorf(`failed parsing current tag %q: %v`, *currentTag, err)
//		}
//		parsedNewTag, err := semver.Parse(newTag)
//		if err != nil {
//			return nil, fmt.Errorf(`failed parsing new tag %q: %v`, newTag, err)
//		}
//		if semver.IsWantedUpgrade(parsedCurrentTag, parsedNewTag, trigger.Image.Track) {
//			updates = append(updates, ImageUpdate{
//				ImageRepo: imageRepo,
//				NewTag:    newTag,
//				TagValue:  tagValue,
//				Release:   release,
//			})
//		}
//	}
//	return updates, nil
//}

func ReleaseWantsChartUpdate(release *model.Release, newVersion string, env *model.Environment) ([]ChartUpdate, error) {
	newSemver, err := semver.Parse(newVersion)
	if err != nil {
		return nil, fmt.Errorf(`release %q: new version %q is not semver`, release.Name, newVersion)
	}
	if release.Chart.Version == nil {
		return nil, fmt.Errorf(`release %q: missing chart.version`, release.Name)
	}
	updates := make([]ChartUpdate, 0)
	for _, trigger := range release.Triggers {
		if trigger.Chart == nil || trigger.Chart.Track == "" {
			continue
		}
		currentSemver, err := semver.Parse(*release.Chart.Version)
		if err != nil {
			continue
		}
		if semver.IsWantedUpgrade(currentSemver, newSemver, trigger.Chart.Track) {
			updates = append(updates, ChartUpdate{
				Release:    release,
				NewVersion: newVersion,
				OldVersion: *release.Chart.Version,
				Reason:     fmt.Sprintf(`track=%q allows upgrade from %q to %q`, trigger.Chart.Track, *release.Chart.Version, newVersion),
			})
		}
	}
	return updates, nil
}

func FindImageUpdatesForRelease(release *model.Release) (map[string][]ImageUpdate, error) {
	updates := make(map[string][]ImageUpdate, 0)
	if release.Triggers == nil {
		return updates, nil
	}
	for _, trigger := range release.Triggers {
		if trigger.Image == nil || trigger.Image.Track == "" {
			continue
		}
		tagValue := trigger.Image.TagValueString()
		repoValue := trigger.Image.RepoValueString()
		prefixValue := trigger.Image.RepoPrefixValueString()
		values, err := helm.GetResolvedValues(release)
		if err != nil {
			return nil, fmt.Errorf(`while looking for updates for release %q: %v`, release.Name, err)
		}
		if helm.KeyIsInValues(repoValue, values) {
			imageRepo := helm.LookupValueByString(repoValue, values).(string)
			if helm.KeyIsInValues(prefixValue, values) {
				imageRepo = helm.LookupValueByString(prefixValue, values).(string) + imageRepo
			}
			imageTag := helm.LookupValueByString(tagValue, values).(string)
			allTags, err := GetTagsForDockerImage(imageRepo)
			if err != nil {
				return nil, fmt.Errorf(`while looking up tags for image %q in release %q: %v`, imageRepo, release.Name, err)
			}
			tagTimestamp := int64(0)
			for _, tsTag := range allTags {
				if tsTag.Tag == imageTag {
					tagTimestamp = tsTag.Timestamp
					break
				}
			}
			updatedTag := GetNewestMatchingTag(imageTag, allTags, trigger.Image.Track, tagTimestamp)
			if updatedTag != "" {
				relFile := release.FromFile
				if _, found := updates[relFile]; !found {
					updates[relFile] = make([]ImageUpdate, 0)
				}
				updates[relFile] = append(updates[relFile], ImageUpdate{
					OldTag:    imageTag,
					NewTag:    updatedTag,
					Release:   release,
					TagValue:  tagValue,
					ImageRepo: imageRepo,
				})
			}
		}
	}
	return updates, nil
}

func FindImageUpdatesForReleases(releases []*model.Release) (map[string][]ImageUpdate, error) {
	envUpdates := make(map[string][]ImageUpdate, 0)
	for _, release := range releases {
		imageUpdates, err := FindImageUpdatesForRelease(release)
		if err != nil {
			return nil, fmt.Errorf(`while looking for updates for release %q in env %q: %v`, release.Name, release.Environment.Name, err)
		}
		for file, updates := range imageUpdates {
			if _, found := envUpdates[file]; !found {
				envUpdates[file] = make([]ImageUpdate, 0)
			}
			for _, update := range updates {
				envUpdates[file] = append(envUpdates[file], update)
			}
		}
	}
	return envUpdates, nil
}

func FindImageUpdatesForEnv(env *model.Environment) (map[string][]ImageUpdate, error) {
	return FindImageUpdatesForReleases(env.AllReleases())
}
