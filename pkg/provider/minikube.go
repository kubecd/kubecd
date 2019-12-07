package provider

import (
	"github.com/kubecd/kubecd/pkg/model"
)

type MinikubeClusterProvider struct{ baseClusterProvider }

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
