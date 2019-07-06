package kubecd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/image"
	"github.com/zedge/kubecd/pkg/model"
)

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

//func ReleaseWantsImageUpdate(release *model.Release, newImage *image.DockerImageRef) ([]ImageUpdate, error) {
//	updates := make([]ImageUpdate, 0)
//	for _, trigger := range release.Triggers {
//		if trigger.Image == nil {
//			//fmt.Printf("no image trigger: release %q", release.Name)
//			continue
//		}
//		values, err := helm.GetResolvedValues(release)
//		if err != nil {
//			return nil, err
//		}
//		imageRepo := helm.GetImageRefFromImageTrigger(trigger.Image, values)
//		if imageRepo.WithoutTag() != newImage.WithoutTag() {
//			//fmt.Printf("imageRepo != newImage: %q != %q", imageRepo, newImage.WithoutTag())
//			continue
//		}
//		if imageRepo.Tag == newImage.Tag {
//			continue
//		}
//		tagValue := trigger.Image.TagValueString()
//		currentTag := helm.LookupValueByString(tagValue, values).(*string)
//		if currentTag == nil {
//			continue
//		}
//		// If the current version is not semver, or track is "Newest", blindly treat any value as an update
//		if !semver.IsSemver(*currentTag) || trigger.Image.Track == semver.TrackNewest {
//			reason := `track=Newest, any observed tag considered newer`
//			if trigger.Image.Track != semver.TrackNewest {
//				reason = fmt.Sprintf(`current tag %q not semver, any observed tag considered newer`, *currentTag)
//			}
//			updates = append(updates, ImageUpdate{
//				ImageRepo: imageRepo.WithoutTag(),
//				NewTag:    newImage.Tag,
//				TagValue:  tagValue,
//				Release:   release,
//				Reason:    reason,
//			})
//			continue
//		}
//		parsedCurrentTag, err := semver.Parse(*currentTag)
//		if err != nil {
//			return nil, fmt.Errorf(`release %q: failed parsing current tag %q: %v`, release.Name, *currentTag, err)
//		}
//		parsedNewTag, err := semver.Parse(newImage.Tag)
//		if err != nil {
//			return nil, fmt.Errorf(`release %q: failed parsing new tag %q: %v`, release.Name, newImage.Tag, err)
//		}
//		if semver.IsWantedUpgrade(parsedCurrentTag, parsedNewTag, trigger.Image.Track) {
//			updates = append(updates, ImageUpdate{
//				ImageRepo: imageRepo.WithoutTag(),
//				NewTag:    newImage.Tag,
//				TagValue:  tagValue,
//				Release:   release,
//				Reason:    fmt.Sprintf(`with track=%q, %q > %q`, trigger.Image.Track, newImage.Tag, *currentTag),
//			})
//		}
//	}
//	return updates, nil
//}

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
//		imageRepo := helm.GetImageRefFromImageTrigger(trigger.Image, values)
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

//func ReleaseWantsChartUpdate(release *model.Release, newVersion string, env *model.Environment) ([]ChartUpdate, error) {
//	newSemver, err := semver.Parse(newVersion)
//	if err != nil {
//		return nil, fmt.Errorf(`release %q: new version %q is not semver`, release.Name, newVersion)
//	}
//	if release.Chart.Version == nil {
//		return nil, fmt.Errorf(`release %q: missing chart.version`, release.Name)
//	}
//	updates := make([]ChartUpdate, 0)
//	for _, trigger := range release.Triggers {
//		if trigger.Chart == nil || trigger.Chart.Track == "" {
//			continue
//		}
//		currentSemver, err := semver.Parse(*release.Chart.Version)
//		if err != nil {
//			continue
//		}
//		if semver.IsWantedUpgrade(currentSemver, newSemver, trigger.Chart.Track) {
//			updates = append(updates, ChartUpdate{
//				Release:    release,
//				NewVersion: newVersion,
//				OldVersion: *release.Chart.Version,
//				Reason:     fmt.Sprintf(`track=%q allows upgrade from %q to %q`, trigger.Chart.Track, *release.Chart.Version, newVersion),
//			})
//		}
//	}
//	return updates, nil
//}

func FindImageUpdatesForRelease(release *model.Release, tagIndex TagIndex) ([]ImageUpdate, error) {
	updates := make([]ImageUpdate, 0)
	if release.Triggers == nil {
		return updates, nil
	}
	for _, trigger := range release.Triggers {
		if trigger.Image == nil || trigger.Image.Track == "" {
			fmt.Println("no trigger")
			continue
		}
		values, err := helm.GetResolvedValues(release)
		if err != nil {
			return nil, fmt.Errorf(`while looking for updates for release %q: %v`, release.Name, err)
		}
		imageRef := helm.GetImageRefFromImageTrigger(trigger.Image, values)
		if imageRef == nil {
			continue
		}
		imageTags := tagIndex.GetTags(imageRef)
		if imageTags == nil {
			//fmt.Printf("env:%s release:%s imageTags == nil, tagIndex=%#v, imageRef=%#v\n", release.Environment.Name, release.Name, tagIndex, *imageRef)
			continue
		}
		var currentTag image.TimestampedTag
		foundTag := false
		for _, tag := range imageTags {
			if imageRef.Tag == tag.Tag {
				currentTag = tag
				foundTag = true
			}
		}
		if !foundTag {
			fmt.Printf("did not find %s in %#v\n", imageRef.Tag, imageTags)
			continue
		}
		newestTag := image.GetNewestMatchingTag(currentTag, imageTags, trigger.Image.Track)
		if newestTag.Tag != currentTag.Tag {
			updates = append(updates, ImageUpdate{
				OldTag: currentTag.Tag,
				NewTag: newestTag.Tag,
				Release: release,
				TagValue: trigger.Image.TagValueString(),
				ImageRepo: imageRef.WithoutTag(),
				Reason: "FIXME",
			})
		}
	}
	return updates, nil
}

//func FindImageUpdatesForReleases(releases []*model.Release) (map[string][]ImageUpdate, error) {
//	envUpdates := make(map[string][]ImageUpdate, 0)
//	for _, release := range releases {
//		imageUpdates, err := FindImageUpdatesForRelease(release)
//		if err != nil {
//			return nil, fmt.Errorf(`while looking for updates for release %q in env %q: %v`, release.Name, release.Environment.Name, err)
//		}
//		for file, updates := range imageUpdates {
//			if _, found := envUpdates[file]; !found {
//				envUpdates[file] = make([]ImageUpdate, 0)
//			}
//			for _, update := range updates {
//				envUpdates[file] = append(envUpdates[file], update)
//			}
//		}
//	}
//	return envUpdates, nil
//}

//func FindImageUpdatesForEnv(env *model.Environment) (map[string][]ImageUpdate, error) {
//	return FindImageUpdatesForReleases(env.AllReleases())
//}

type ReleaseFilterFunc func(*model.Release) bool

func ImageReleaseIndex(kcdConfig *model.KubeCDConfig, filters ...ReleaseFilterFunc) (map[string][]*model.Release, error) {
	result := make(map[string][]*model.Release)
releaseLoop:
	for _, release := range kcdConfig.AllReleases() {
		for _, filter := range filters {
			if !filter(release) {
				continue releaseLoop
			}
		}
		//fmt.Printf("evaluating release %q\n", release.Name)
		values, err := helm.GetResolvedValues(release)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving values for env %q release %q", release.Environment.Name, release.Name)
		}
		//fmt.Printf("release %q triggers: %#v\n", release.Name, release.Triggers)
		for _, t := range release.Triggers {
			if t.Image == nil {
				//fmt.Printf("release %q has no trigger\n", release.Name)
				continue
			}
			repo := helm.GetImageRefFromImageTrigger(t.Image, values).WithoutTag()
			//fmt.Printf("release %q repo: %q\n", release.Name, repo)
			if _, found := result[repo]; !found {
				result[repo] = make([]*model.Release, 0)
			}
			result[repo] = append(result[repo], release)
		}
	}
	return result, nil
}

// BuildTagIndexFromNewImageRef builds a tag index from an image index, with all
// tags being used from the input image.
func BuildTagIndexFromNewImageRef(newImageRef *image.DockerImageRef, imageIndex map[string][]*model.Release) TagIndex {
	imageRepo := newImageRef.WithoutTag()
	tagIndex := TagIndex(make(map[string][]image.TimestampedTag))
	if _, found := imageIndex[imageRepo]; !found {
		return tagIndex
	}
	// Hardcoding Timestamp to 1 here (vs 0 for tags added below) will force
	// FindImageUpdatesForRelease to choose newImageRef's tag over any existing
	// ones for track=Newest, which is the behaviour we want for "kcd observe".
	tagIndex[imageRepo] = []image.TimestampedTag{{Tag: newImageRef.Tag, Timestamp: int64(1)}}
	for _, release := range imageIndex[imageRepo] {
		for _, trigger := range release.Triggers {
			if trigger.Image == nil || trigger.Image.Track == "" {
				fmt.Println("no trigger")
				continue
			}
			values, err := helm.GetResolvedValues(release)
			if err != nil {
				panic(fmt.Sprintf(`resolving values for release %q: %v`, release.Name, err))
			}
			imageRef := helm.GetImageRefFromImageTrigger(trigger.Image, values)
			if imageRef == nil {
				continue
			}
			if imageRef.WithoutTag() != imageRepo {
				panic(fmt.Sprintf(`expected %q == %q`, imageRef.WithoutTag(), imageRepo))
			}
			tagIndex[imageRepo] = append(tagIndex[imageRepo], image.TimestampedTag{Tag: imageRef.Tag, Timestamp: int64(0)})
		}
	}
	return tagIndex
}

// TagIndex maps image repos (without tag) to a list of tags with timestamps
type TagIndex map[string][]image.TimestampedTag

func BuildTagIndexFromDockerRegistries(imageIndex map[string][]*model.Release) (TagIndex, error) {
	tagIndex := TagIndex(make(map[string][]image.TimestampedTag))
	for repo := range imageIndex {
		tags, err := image.GetTagsForDockerImage(repo)
		if err != nil {
			return nil, errors.Wrapf(err, `while scanning tags for %s`, repo)
		}
		tagIndex[repo] = tags
	}
	return tagIndex, nil
}

func (i TagIndex) GetTagTimestamp(imageRef *image.DockerImageRef) int64 {
	if tags, found := i[imageRef.WithoutTag()]; found {
		for _, tsTag := range tags {
			if tsTag.Tag == imageRef.Tag {
				return tsTag.Timestamp
			}
		}
	}
	return int64(0)
}

func (i TagIndex) GetTags(imageRef *image.DockerImageRef) []image.TimestampedTag {
	return i[imageRef.WithoutTag()]
}
