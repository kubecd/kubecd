package provider

import (
	"github.com/zedge/kubecd/pkg/model"
)

type DockerForDesktopClusterProvider struct{ baseClusterProvider }

var _ ClusterProvider = &DockerForDesktopClusterProvider{}

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
