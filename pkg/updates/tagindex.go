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

package updates

import (
	"fmt"
	"github.com/kubecd/kubecd/pkg/helm"
	"github.com/kubecd/kubecd/pkg/image"
	"github.com/kubecd/kubecd/pkg/model"
	"github.com/pkg/errors"
)

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
