package provider

import (
	"github.com/kubecd/kubecd/pkg/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetClusterProvider(t *testing.T) {
	type testCase struct {
		name                 string
		cluster              *model.Cluster
		expectedProviderType interface{}
	}
	for _, tc := range []testCase{
		{"gke", &model.Cluster{Provider: model.Provider{GKE: &model.GkeProvider{}}}, &GkeClusterProvider{}},
		{"aks", &model.Cluster{Provider: model.Provider{AKS: &model.AksProvider{}}}, &AksClusterProvider{}},
		{"docker", &model.Cluster{Provider: model.Provider{DockerForDesktop: &model.DockerForDesktopProvider{}}}, &DockerForDesktopClusterProvider{}},
		{"minikube", &model.Cluster{Provider: model.Provider{Minikube: &model.MinikubeProvider{}}}, &MinikubeClusterProvider{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cp, err := GetClusterProvider(tc.cluster, false)
			assert.NoError(t, err)
			assert.IsType(t, tc.expectedProviderType, cp)
			cp, err = GetClusterProvider(tc.cluster, true)
			assert.NoError(t, err)
			assert.IsType(t, &GitlabClusterProvider{}, cp)
		})
	}
}
func TestGetContextInitCommands(t *testing.T) {
	env := &model.Environment{Name: "test", KubeNamespace: "default"}
	minikube := &MinikubeClusterProvider{}
	cmds := GetContextInitCommands(minikube, env)
	assert.Equal(t, [][]string{{"kubectl", "config", "set-context", "env:test", "--cluster", "minikube", "--user", "minikube", "--namespace", "default"}}, cmds)
}
