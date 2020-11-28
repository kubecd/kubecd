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

// Package model :
package model

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io"
	"io/ioutil"
	"os"
)

type NameFromRef struct {
	ClusterParam string `json:"clusterParam"`
}

type KubeCDConfig struct {
	Clusters     []*Cluster     `json:"clusters"`
	Environments []*Environment `json:"environments"`
	HelmRepos    []HelmRepo     `json:"helmRepos,omitempty"`
	KubeConfig   *string        `json:"kubeConfig,omitempty"`

	fromFile string
}

func NewConfigFromFile(fileName string) (*KubeCDConfig, error) {
	r, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error while opening %s: %v", fileName, err)
	}
	return NewConfig(r, fileName)
}

func NewConfig(reader io.Reader, fromFile string) (*KubeCDConfig, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error while reading %s: %v", fromFile, err)
	}
	config := &KubeCDConfig{fromFile: fromFile}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling Release from %s: %v", fromFile, err)
	}
	for _, env := range config.Environments {
		env.Cluster = config.GetCluster(env.ClusterName)
		env.fromFile = fromFile
		if env.Cluster == nil {
			return nil, fmt.Errorf(`environment %q refers to undefined Cluster %q`, env.Name, env.ClusterName)
		}
		if err = env.populateReleases(); err != nil {
			return nil, err
		}
	}
	if errs := config.sanityCheck(); errs != nil {
		return nil, NewAggregateError(errs)
	}
	return config, nil
}

func (k *KubeCDConfig) sanityCheck() []error {
	var issues []error
	seenCluster := make(map[string]bool)
	for _, cluster := range k.Clusters {
		if _, seen := seenCluster[cluster.Name]; seen {
			issues = append(issues, fmt.Errorf(`duplicate cluster name: %q`, cluster.Name))
		}
		seenCluster[cluster.Name] = true
		issues = append(issues, cluster.sanityCheck()...)
	}
	seenEnv := make(map[string]bool)
	for _, env := range k.Environments {
		if _, seen := seenEnv[env.Name]; seen {
			issues = append(issues, fmt.Errorf(`duplicate environment name: %q`, env.Name))
		}
		seenEnv[env.Name] = true
		issues = append(issues, env.sanityCheck()...)
	}
	seenHelmRepo := make(map[string]bool)
	for _, repo := range k.HelmRepos {
		if _, seen := seenHelmRepo[repo.Name]; seen {
			issues = append(issues, fmt.Errorf(`duplicate helm repo name: %q`, repo.Name))
		}
		seenHelmRepo[repo.Name] = true
	}
	return issues
}

func (k *KubeCDConfig) AllClusters() []*Cluster {
	return k.Clusters
}

func (k *KubeCDConfig) AllReleases() []*Release {
	result := make([]*Release, 0)
	for _, env := range k.Environments {
		result = append(result, env.Releases...)
	}
	return result
}

func (k *KubeCDConfig) GetCluster(name string) *Cluster {
	for _, cluster := range k.Clusters {
		if cluster.Name == name {
			return cluster
		}
	}
	return nil
}

func (k *KubeCDConfig) GetEnvironment(name string) *Environment {
	for _, env := range k.Environments {
		if env.Name == name {
			return env
		}
	}
	return nil
}

func (k *KubeCDConfig) GetEnvironmentsInCluster(clusterName string) []*Environment {
	var envs []*Environment
	for _, env := range k.Environments {
		if env.ClusterName == clusterName {
			envs = append(envs, env)
		}
	}
	return envs
}

func (k *KubeCDConfig) HasCluster(name string) bool {
	cluster := k.GetCluster(name)
	return cluster != nil
}

func (k *KubeCDConfig) FromFile() string {
	return k.fromFile
}

func KubeContextName(envName string) string {
	return "env:" + envName
}
