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
