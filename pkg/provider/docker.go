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
	"github.com/kubecd/kubecd/pkg/model"
)

type DockerForDesktopClusterProvider struct{ baseClusterProvider }

func (p *DockerForDesktopClusterProvider) GetClusterInitCommands() ([][]string, error) {
	return [][]string{{
		"kubectl",
		"config",
		"set-cluster",
		"docker-for-desktop-cluster",
		"--insecure-skip-tls-verify=true",
		"--server=https://localhost:6443",
	}}, nil
}

func (p *DockerForDesktopClusterProvider) GetClusterName() string {
	return "docker-for-desktop-cluster"
}

func (p *DockerForDesktopClusterProvider) GetUserName() string {
	return "docker-for-desktop"
}

func (p *DockerForDesktopClusterProvider) GetNamespace(env *model.Environment) string {
	return env.KubeNamespace
}
