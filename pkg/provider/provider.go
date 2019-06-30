package provider

import (
	"fmt"
	"github.com/zedge/kubecd/pkg/model"
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
