package provider

import (
	"fmt"
	"github.com/kubecd/kubecd/pkg/model"
	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
)

type ExistingContextClusterProvider struct {
	baseClusterProvider
	context *clientapi.Context
}

func NewExistingContextClusterProvider(provider baseClusterProvider) (*ExistingContextClusterProvider, error) {
	kubeConfigPath, err := findKubeConfig()
	if err != nil {
		return nil, err
	}

	kubeConfig, err := clientcmd.LoadFromFile(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	contextName := provider.Cluster.Provider.ExistingContext.ContextName
	if context, found := kubeConfig.Contexts[contextName]; found {
		return &ExistingContextClusterProvider{
			baseClusterProvider: provider,
			context:             context,
		}, nil
	}
	return nil, fmt.Errorf("context %q not found in kubeconfig", contextName)
}

func (p *ExistingContextClusterProvider) GetClusterInitCommands() ([][]string, error) {
	return [][]string{{}}, nil
}

func (p *ExistingContextClusterProvider) GetClusterName() string {
	return p.context.Cluster
}

func (p *ExistingContextClusterProvider) GetUserName() string {
	return p.context.AuthInfo
}

func (p *ExistingContextClusterProvider) GetNamespace(env *model.Environment) string {
	return env.KubeNamespace
}

func findKubeConfig() (string, error) {
	env := os.Getenv("KUBECONFIG")
	if env != "" {
		return env, nil
	}
	path, err := homedir.Expand("~/.kube/config")
	if err != nil {
		return "", err
	}
	return path, nil
}
