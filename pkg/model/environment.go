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
		releaseList, err := NewReleaseListFromFile(ResolvePathFromFile(releaseListFile, e.fromFile))
		if err != nil {
			return fmt.Errorf(`environment %q: %v`, e.Name, err)
		}
		for _, rel := range releaseList.Releases {
			e.Releases = append(e.Releases, rel)
		}
	}

	return nil
}

func (e *Environment) sanityCheck() []error {
	var issues []error
	seenEnv := make(map[string]bool)
	for _, rel := range e.Releases {
		if _, seen := seenEnv[rel.Name]; seen {
			issues = append(issues, fmt.Errorf(`environment %q: duplicate release name %q`, e.Name, rel.Name))
		}
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
