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

package image

import (
	"encoding/json"
	"fmt"
	mmsemver "github.com/Masterminds/semver"
	registry2 "github.com/heroku/docker-registry-client/registry"
	"github.com/kubecd/kubecd/pkg/semver"
	"strings"
	"time"
)

const (
	DefaultDockerRegistry = "registry-1.docker.io"
	GCRRegistrySuffix     = "gcr.io"
)

func ParseDockerTimestamp(str string) (int64, error) {
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

func (r DockerImageRef) RegistryURL() string {
	return "https://" + r.Registry
}

func (r DockerImageRef) WithTag() string {
	return r.WithoutTag() + ":" + r.Tag
}

func (r DockerImageRef) WithoutTag() string {
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

type TimestampedTag struct {
	Tag       string
	Timestamp int64
	semantic  *mmsemver.Version
}

func (t *TimestampedTag) Semantic() *mmsemver.Version {
	if t.semantic == nil {
		t.semantic, _ = semver.Parse(t.Tag)
	}
	return t.semantic
}

func GetTagsForGcrImage(repo *DockerImageRef) ([]TimestampedTag, error) {
	fullRepo := repo.WithoutTag()
	tmp := strings.Split(repo.Image, "/")
	gcpProject := tmp[0]
	output, err := runner.Run("gcloud", "container", "images", "list-tags", fullRepo, "--project", gcpProject, "--format", "json")
	if err != nil {
		return nil, fmt.Errorf(`failed listing tags for GCR image %q: %v`, fullRepo, err)
	}
	result := make([]TimestampedTag, 0)
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
			result = append(result, TimestampedTag{Tag: tag, Timestamp: timestamp.Unix()})
		}
	}
	return result, nil
}

func GetTagsForDockerHubImage(repo *DockerImageRef) ([]TimestampedTag, error) {
	return GetTagsForDockerV2RegistryImage(repo)
}

func GetTagsForDockerV2RegistryImage(repo *DockerImageRef) ([]TimestampedTag, error) {
	registry, err := registry2.New(repo.RegistryURL(), "", "")
	result := make([]TimestampedTag, 0)
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
			timestamp, err := ParseDockerTimestamp(v1Compat.Created)
			if err != nil {
				return nil, fmt.Errorf(`could not parse timestamp for %s:%s: %v`, repo.WithoutTag(), tag, err)
			}
			result = append(result, TimestampedTag{Tag: tag, Timestamp: timestamp})
		}
	}
	return result, nil
}

func GetTagsForDockerImage(repo string) ([]TimestampedTag, error) {
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

// GetNewestMatchingTag returns the "newest" (as defined by the track parameter) candidate
// tag, or the current tag if no better ones were found.
func GetNewestMatchingTag(currentTag TimestampedTag, candidateTags []TimestampedTag, track string) TimestampedTag {
	var foundTag = currentTag
	if track == semver.TrackNewest {
		for _, candidateTag := range candidateTags {
			if candidateTag.Tag == "latest" {
				continue
			}
			if candidateTag.Timestamp > foundTag.Timestamp {
				foundTag = candidateTag
			}
		}
	}
	semanticTags := make([]*mmsemver.Version, 0)
	semTagMap := make(map[string]TimestampedTag)
	for _, ct := range candidateTags {
		if st := ct.Semantic(); st != nil {
			semanticTags = append(semanticTags, st)
			semTagMap[st.String()] = ct
		}
	}
	newest, err := semver.BestUpgrade(currentTag.Semantic(), semanticTags, track)
	if err != nil {
		return foundTag
	}
	return semTagMap[newest.String()]
}
