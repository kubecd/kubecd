package provider

import (
	"github.com/zedge/kubecd/pkg/model"
)

type MinikubeClusterProvider struct { baseClusterProvider }

var _ ClusterProvider = &MinikubeClusterProvider{}

func (p *MinikubeClusterProvider) GetClusterInitCommands() ([][]string, error) {
	return [][]string{}, nil
}

func (p *MinikubeClusterProvider) GetClusterName() string {
	return "minikube"
}

func (p *MinikubeClusterProvider) GetUserName() string {
	return "minikube"
}

func (p *MinikubeClusterProvider) GetNamespace(env *model.Environment) string {
	return env.KubeNamespace
}
