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

package model

const (
	DefaultTagValue        = "image.tag"
	DefaultRepoValue       = "image.repository"
	DefaultRepoPrefixValue = "image.prefix"
)

type ImageTrigger struct {
	TagValue        string `json:"tagValue"`
	RepoValue       string `json:"repoValue"`
	RepoPrefixValue string `json:"repoPrefixValue"`
	Track           string `json:"track"` // one of "PatchLevel", "MinorVersion", "MajorVersion", "Newest"
}

func (t *ImageTrigger) TagValueString() string {
	if t.TagValue == "" {
		return DefaultTagValue
	}
	return t.TagValue
}

func (t *ImageTrigger) RepoValueString() string {
	if t.RepoValue == "" {
		return DefaultRepoValue
	}
	return t.RepoValue
}

func (t *ImageTrigger) RepoPrefixValueString() string {
	if t.RepoPrefixValue == "" {
		return DefaultRepoPrefixValue
	}
	return t.RepoPrefixValue
}

type HelmTrigger struct {
	Track string `json:"track"` // one of "PatchLevel", "MinorVersion", "MajorVersion", "Newest"
}

type ReleaseUpdateTrigger struct {
	Image *ImageTrigger `json:"image,omitempty"`
	Chart *HelmTrigger  `json:"chart,omitempty"`
}
