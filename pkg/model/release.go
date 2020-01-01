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

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
)

type Release struct {
	Name              string                 `json:"name"`
	Chart             *Chart                 `json:"chart,omitempty"`
	ValuesFile        *string                `json:"valuesFile,omitempty"`
	Values            []ChartValue           `json:"values,omitempty"`
	Trigger           *ReleaseUpdateTrigger  `json:"trigger,omitempty"`
	Triggers          []ReleaseUpdateTrigger `json:"triggers,omitempty"`
	SkipDefaultValues bool                   `json:"skipDefaultValues,omitempty"`
	ResourceFiles     []string               `json:"resourceFiles,omitempty"`

	FromFile    string       `json:"-"`
	Environment *Environment `json:"-"`
}

type ReleaseList struct {
	ResourceFiles []string   `json:"resourceFiles,omitempty"`
	Releases      []*Release `json:"releases,omitempty"`

	FromFile string
}

//func NewRelease(reader io.Reader, FromFile string) (*Release, error) {
//	data, err := ioutil.ReadAll(reader)
//	if err != nil {
//		return nil, fmt.Errorf("error while reading %s: %v", FromFile, err)
//	}
//	release := &Release{FromFile: FromFile}
//	err = yaml.Unmarshal(data, release)
//	if err != nil {
//		return nil, fmt.Errorf("error while unmarshaling Release from %s: %v", FromFile, err)
//	}
//	if issues := release.sanityCheck(); len(issues) > 0 {
//		return nil, fmt.Errorf("issues found in Release %q: %v", release.Name, issues)
//	}
//	return release, nil
//}

func NewReleaseListFromFile(env *Environment, fromFile string) (*ReleaseList, error) {
	r, err := os.Open(fromFile)
	if err != nil {
		return nil, fmt.Errorf("error while opening %s: %v", fromFile, err)
	}
	return NewReleaseList(env, r, fromFile)
}

func NewReleaseList(env *Environment, reader io.Reader, fromFile string) (*ReleaseList, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error while reading %s: %v", fromFile, err)
	}
	releaseList := &ReleaseList{FromFile: fromFile}
	err = yaml.Unmarshal(data, releaseList)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling Release from %s: %v", fromFile, err)
	}
	for _, release := range releaseList.Releases {
		release.FromFile = fromFile
		release.Environment = env
	}
	return releaseList, nil
}

func (r *Release) sanityCheck() []error {
	var issues []error
	if r.Chart == nil && (r.ResourceFiles == nil || len(r.ResourceFiles) == 0) {
		issues = append(issues, fmt.Errorf(`release %q: must define either "chart" or "resourceFiles"`, r.Name))
	}
	if r.Chart != nil {
		if r.ResourceFiles != nil && len(r.ResourceFiles) > 0 {
			issues = append(issues, fmt.Errorf(`release %q: must define only one of "chart" or "resourceFiles"`, r.Name))
		}
		if r.Chart.Reference != nil && r.Chart.Version == nil {
			issues = append(issues, fmt.Errorf(`release %q: must have a chart.version`, r.Name))
		}
	}
	return issues
}

func (r *Release) AbsPath(path string) string {
	return ResolvePathFromFile(path, r.FromFile)
}

func (l *ReleaseList) sanityCheck() []error {
	var issues []error
	for _, rel := range l.Releases {
		for _, issue := range rel.sanityCheck() {
			issues = append(issues, issue)
		}
	}
	return issues
}

func (r *Release) UnmarshalJSON(data []byte) error {
	type release Release
	if err := json.Unmarshal(data, (*release)(r)); err != nil {
		return err
	}
	if r.Trigger != nil {
		r.Triggers = append(r.Triggers, *r.Trigger)
		r.Trigger = nil
	}
	return nil
}
