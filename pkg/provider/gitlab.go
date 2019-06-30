package provider

import "github.com/zedge/kubecd/pkg/model"

type GitlabClusterProvider struct { baseClusterProvider }

var _ ClusterProvider = &GitlabClusterProvider{}

func (p *GitlabClusterProvider) GetClusterName() string {
	return "gitlab-deploy"
}

func (p *GitlabClusterProvider) GetUserName() string {
	return "gitlab-deploy"
}

func (p *GitlabClusterProvider) GetNamespace(env *model.Environment) string {
	return env.KubeNamespace
}

func (p *GitlabClusterProvider) GetClusterInitCommands() ([][]string, error) {
	return [][]string{}, nil
}
