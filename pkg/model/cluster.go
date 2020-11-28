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
	"github.com/pkg/errors"
)

type ClusterParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cluster struct {
	Name       string             `json:"name"`
	Provider   Provider           `json:"provider"`
	Parameters []ClusterParameter `json:"parameters"`
}

type GceAddressValueRef struct {
	Name     string      `json:"name,omitempty"`
	NameFrom NameFromRef `json:"nameFrom,omitempty"`
	IsGlobal bool        `json:"isGlobal,omitempty"` // if false, use zone/region from Cluster
}

type GceValueRef struct {
	Address *GceAddressValueRef `json:"address,omitempty"`
}

func (c *Cluster) sanityCheck() []error {
	var issues []error
	providers := 0
	if c.Provider.GKE != nil {
		providers++
		issues = append(issues, c.Provider.GKE.sanityCheck()...)
	}
	if c.Provider.Minikube != nil {
		providers++
	}
	if c.Provider.AKS != nil {
		providers++
	}
	if c.Provider.DockerForDesktop != nil {
		providers++
	}
	if c.Provider.ExistingContext != nil {
		providers++
	}
	if providers != 1 {
		issues = append(issues, errors.New(`must specify one and ony one provider`))
	}
	return issues
}

func (p *GkeProvider) sanityCheck() []error {
	var issues []error
	tmp := 0
	if p.Zone != nil {
		tmp++
	}
	if p.Region != nil {
		tmp++
	}
	if tmp != 1 {
		issues = append(issues, fmt.Errorf(`must specify either "zone" or "region for GKE Cluster %q"`, p.ClusterName))
	}
	return issues
}
