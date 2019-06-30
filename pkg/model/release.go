package model

import (
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

	fromFile string
}

type ReleaseList struct {
	ResourceFiles []string  `json:"resourceFiles,omitempty"`
	Releases      []*Release `json:"releases,omitempty"`

	fromFile string
}

func NewRelease(reader io.Reader, fromFile string) (*Release, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error while reading %s: %v", fromFile, err)
	}
	release := &Release{fromFile: fromFile}
	err = yaml.Unmarshal(data, release)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling Release from %s: %v", fromFile, err)
	}
	if issues := release.sanityCheck(); len(issues) > 0 {
		return nil, fmt.Errorf("issues found in Release %q: %v", release.Name, issues)
	}
	return release, nil
}

func NewReleaseListFromFile(fromFile string ) (*ReleaseList, error) {
	r, err := os.Open(fromFile)
	if err != nil {
		return nil, fmt.Errorf("error while opening %s: %v",fromFile, err)
	}
	return NewReleaseList(r, fromFile)
}

func NewReleaseList(reader io.Reader, fromFile string) (*ReleaseList, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error while reading %s: %v", fromFile, err)
	}
	releaseList := &ReleaseList{fromFile: fromFile}
	err = yaml.Unmarshal(data, releaseList)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling Release from %s: %v", fromFile, err)
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

func (r *Release) FromFile() string {
	return r.fromFile
}

func (r *Release) AbsPath(path string) string {
	return ResolvePathFromFile(path, r.fromFile)
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
