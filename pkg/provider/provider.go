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

package provider

import (
	"fmt"
	"github.com/kubecd/kubecd/pkg/model"
)

type ClusterProvider interface {
	GetClusterInitCommands() ([][]string, error)
	GetClusterName() string
	GetUserName() string
	GetNamespace(environment *model.Environment) string
}

type baseClusterProvider struct {
	*model.Cluster
}

func GetClusterProvider(cluster *model.Cluster, gitlabMode bool) (ClusterProvider, error) {
	var provider ClusterProvider
	switch {
	case gitlabMode:
		provider = &GitlabClusterProvider{baseClusterProvider{cluster}}
	case cluster.Provider.GKE != nil:
		provider = &GkeClusterProvider{baseClusterProvider{cluster}}
	case cluster.Provider.AKS != nil:
		provider = &AksClusterProvider{baseClusterProvider{cluster}}
	case cluster.Provider.Minikube != nil:
		provider = &MinikubeClusterProvider{baseClusterProvider{cluster}}
	case cluster.Provider.DockerForDesktop != nil:
		provider = &DockerForDesktopClusterProvider{baseClusterProvider{cluster}}
	default:
		return nil, fmt.Errorf(`could not find a provider for cluster %q`, cluster.Name)
	}
	return provider, nil
}

func GetContextInitCommands(provider ClusterProvider, env *model.Environment) [][]string {
	return [][]string{{
		"kubectl", "config", "set-context", model.KubeContextName(env.Name),
		"--cluster", provider.GetClusterName(),
		"--user", provider.GetUserName(),
		"--namespace", provider.GetNamespace(env),
	}}
}
