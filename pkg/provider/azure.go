package provider

import (
	"github.com/zedge/kubecd/pkg/model"
)

type AksClusterProvider struct{ baseClusterProvider }

func (p *AksClusterProvider) GetClusterInitCommands() ([][]string, error) {
	panic("implement me")
}

func (p *AksClusterProvider) GetClusterName() string {
	panic("implement me")
}

func (p *AksClusterProvider) GetUserName() string {
	panic("implement me")
}

func (p *AksClusterProvider) GetNamespace(env *model.Environment) string {
	panic("implement me")
}
