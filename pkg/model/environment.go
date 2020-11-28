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
package model

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io"
	"io/ioutil"
)

type Environment struct {
	Name              string       `json:"name"`
	ClusterName       string       `json:"clusterName"`
	Namespace         string       `json:"namespace"`
	KubeNamespace     string       `json:"kubeNamespace"`
	ReleasesFiles     []string     `json:"releasesFiles,omitempty"`
	DefaultValuesFile string       `json:"defaultValuesFile,omitempty"`
	DefaultValues     []ChartValue `json:"defaultValues,omitempty"`
	Releases          []*Release   `json:"releases,omitempty"`
	Cluster           *Cluster     `json:"-"`

	fromFile string
}

func NewEnvironment(reader io.Reader, envFile string) (*Environment, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error while reading %s: %v", envFile, err)
	}
	env := &Environment{fromFile: envFile}
	err = yaml.Unmarshal(data, env)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling Environment from %s: %v", envFile, err)
	}
	if err = env.populateReleases(); err != nil {
		return nil, err
	}
	if issues := env.sanityCheck(); len(issues) > 0 {
		return nil, fmt.Errorf("issues found in Environment %q: %v", env.Name, issues)
	}
	return env, nil
}

func (e *Environment) populateReleases() error {
	for _, releaseListFile := range e.ReleasesFiles {
		releaseList, err := NewReleaseListFromFile(e, ResolvePathFromFile(releaseListFile, e.fromFile))
		if err != nil {
			return fmt.Errorf(`environment %q: %v`, e.Name, err)
		}
		e.Releases = append(e.Releases, releaseList.Releases...)
	}

	return nil
}

func (e *Environment) sanityCheck() []error {
	var issues []error
	seenRelease := make(map[string]bool)
	for _, rel := range e.Releases {
		if _, seen := seenRelease[rel.Name]; seen {
			issues = append(issues, fmt.Errorf(`duplicate release %q in environment %q`, rel.Name, e.Name))
		}
		seenRelease[rel.Name] = true
		issues = append(issues, rel.sanityCheck()...)
	}
	return issues
}

func (e *Environment) AllReleases() []*Release {
	return e.Releases
}

func (e *Environment) GetRelease(name string) *Release {
	for _, rel := range e.Releases {
		if rel.Name == name {
			return rel
		}
	}
	return nil
}

func (e *Environment) GetCluster() *Cluster {
	return e.Cluster
}
