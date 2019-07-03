package model

import (
	"errors"
	"fmt"
)

type ClusterParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cluster struct {
	Name       string             `json:"name"`
	Provider   Provider           `json:"provider"`
	Parameters []ClusterParameter `json:"parameters"`

	fromFile string
}

type GceAddressValueRef struct {
	Name     string      `json:"name,optional"`
	NameFrom NameFromRef `json:"nameFrom,optional"`
	IsGlobal bool        `json:"isGlobal,optional"` // if false, use zone/region from Cluster
}

type GceValueRef struct {
	Address *GceAddressValueRef `json:"address,optional"`
}

func (c *Cluster) sanityCheck() []error {
	var issues []error
	providers := 0
	if c.Provider.GKE != nil {
		providers++
		for _, issue := range c.Provider.GKE.sanityCheck() {
			issues = append(issues, issue)
		}
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
