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

type GkeClusterProvider struct{ baseClusterProvider }

func (p *GkeClusterProvider) GetClusterInitCommands() ([][]string, error) {
	gcloudCommand := []string{
		"gcloud", "container", "clusters", "get-credentials", "--project", p.Provider.GKE.Project,
	}
	if p.Provider.GKE.Zone != nil {
		gcloudCommand = append(gcloudCommand, "--zone", *p.Provider.GKE.Zone)
	} else {
		gcloudCommand = append(gcloudCommand, "--region", *p.Provider.GKE.Region)
	}
	gcloudCommand = append(gcloudCommand, p.Provider.GKE.ClusterName)
	return [][]string{gcloudCommand}, nil
}

func (p *GkeClusterProvider) GetClusterName() string {
	// 'gke_{gke.project}_{zone_or_region}_{gke.clusterName}'
	return fmt.Sprintf("gke_%s_%s_%s", p.Provider.GKE.Project, regionOrZone(p.Provider.GKE), p.Provider.GKE.ClusterName)
}

func regionOrZone(gke *model.GkeProvider) string {
	if gke.Region != nil {
		return *gke.Region
	}
	return *gke.Zone
}

func (p *GkeClusterProvider) GetUserName() string {
	gke := p.Provider.GKE
	return fmt.Sprintf("gke_%s_%s_%s", gke.Project, regionOrZone(gke), gke.ClusterName)
}

func (p *GkeClusterProvider) GetNamespace(environment *model.Environment) string {
	return environment.KubeNamespace
}
