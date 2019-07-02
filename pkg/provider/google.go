package provider

import (
	"fmt"
	"github.com/zedge/kubecd/pkg/model"
)

type GkeClusterProvider struct { baseClusterProvider }

var _ ClusterProvider = &GkeClusterProvider{}

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
