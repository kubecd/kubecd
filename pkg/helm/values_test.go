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

package helm

import (
	"fmt"
	"github.com/kubecd/kubecd/pkg/exec"
	"github.com/kubecd/kubecd/pkg/image"
	"github.com/kubecd/kubecd/pkg/model"
	"github.com/kubecd/kubecd/pkg/semver"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

const testIpAddress = "1.2.3.4"

func TestLookupValue(t *testing.T) {
	values := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": "baz",
		},
		"very": map[string]interface{}{
			"very": map[string]interface{}{
				"very": map[string]interface{}{
					"very": "deep",
				},
			},
		},
		"a": "b",
	}
	for key, expectedResult := range map[string]interface{}{
		"foo":                      nil,
		"foo.bar":                  "baz",
		"very":                     nil,
		"very.very":                nil,
		"very.very.very":           nil,
		"very.very.very.very":      "deep",
		"very.very.very.very.deep": nil,
		"unknown":                  nil,
		"a":                        "b",
	} {
		result := LookupValueByString(key, values)
		if expectedResult == nil {
			assert.Nil(t, result)
		} else {
			assert.Equal(t, expectedResult, *result.(*string))
		}
	}
}

func TestResolveGceAddressValue(t *testing.T) {
	oldRunner := runner
	defer func() { runner = oldRunner }()
	runner = exec.TestRunner{Output: []byte(testIpAddress)}
	zone := "us-central1-a"
	cluster := model.Cluster{
		Name: "kcd-clustername",
		Provider: model.Provider{
			GKE: &model.GkeProvider{
				Project:     "test-project",
				Zone:        &zone,
				ClusterName: "gke-clustername",
			},
		},
	}
	env := &model.Environment{
		Cluster: &cluster,
	}
	address := &model.GceAddressValueRef{
		Name:     "my-address",
		IsGlobal: false,
	}
	out, err := ResolveGceAddressValue(address, env)
	assert.NoError(t, err)
	assert.Equal(t, testIpAddress, string(out))
}

// TestHelperProcess is required boilerplate (one per package) for using exec.TestRunner
func TestHelperProcess(t *testing.T) {
	exec.InsideHelperProcess()
}

func TestGenerateHelmApplyArgv(t *testing.T) {
	chartRef := "stable/cert-manager"
	chartVer := "v0.5.1"
	valuesFile := "values-certmanager.yaml"
	releaseFile := path.Join(os.TempDir(), "releases.yaml")
	expectedValuesFile := path.Join(os.TempDir(), valuesFile)
	releaseName := "cert-manager"
	envName := "kube-system"
	envNamespace := "kube-system"
	release := &model.Release{
		Name: releaseName,
		Chart: &model.Chart{
			Reference: &chartRef,
			Version:   &chartVer,
		},
		ValuesFile: &valuesFile,
		Triggers: []model.ReleaseUpdateTrigger{
			{Chart: &model.HelmTrigger{Track: semver.TrackMinorVersion}},
		},
		FromFile: releaseFile,
	}
	env := &model.Environment{
		Name:          envName,
		KubeNamespace: envNamespace,
	}
	t.Run("release values file only", func(t *testing.T) {
		cmds, err := GenerateHelmApplyArgv(release, env, false, false)
		assert.NoError(t, err)
		assert.Equal(t,
			[]string{
				"helm", "--kube-context", "env:" + envName, "upgrade", releaseName,
				chartRef, "--version", chartVer, "-i", "--namespace", envNamespace,
				"--values", expectedValuesFile},
			cmds)

	})
	t.Run("env and release values files", func(t *testing.T) {
		env.DefaultValuesFile = path.Join(os.TempDir(), "env-values.yaml")
		cmds, err := GenerateHelmApplyArgv(release, env, false, false)
		assert.NoError(t, err)
		assert.Equal(t,
			[]string{
				"helm", "--kube-context", "env:" + envName, "upgrade", releaseName,
				chartRef, "--version", chartVer, "-i", "--namespace", envNamespace,
				"--values", env.DefaultValuesFile, "--values", expectedValuesFile},
			cmds)
	})
	t.Run("env values and release values file", func(t *testing.T) {
		env.DefaultValuesFile = ""
		env.DefaultValues = []model.ChartValue{{Key: "foo", Value: "bar"}}
		cmds, err := GenerateHelmApplyArgv(release, env, false, false)
		assert.NoError(t, err)
		assert.Equal(t,
			[]string{
				"helm", "--kube-context", "env:" + envName, "upgrade", releaseName,
				chartRef, "--version", chartVer, "-i", "--namespace", envNamespace,
				"--set-string", "foo=bar", "--values", expectedValuesFile},
			cmds)
	})
	t.Run("release values file and values", func(t *testing.T) {
		env.DefaultValues = nil
		release.Values = []model.ChartValue{{Key: "baz", Value: "gazonk"}}
		cmds, err := GenerateHelmApplyArgv(release, env, false, false)
		assert.NoError(t, err)
		assert.Equal(t,
			[]string{
				"helm", "--kube-context", "env:" + envName, "upgrade", releaseName,
				chartRef, "--version", chartVer, "-i", "--namespace", envNamespace,
				"--values", expectedValuesFile, "--set-string", "baz=gazonk"},
			cmds)
	})
}

func TestGetImageRepoFromImageTrigger(t *testing.T) {
	trigger := &model.ImageTrigger{}
	valuesWithoutPrefix := map[string]interface{}{
		"image": map[string]interface{}{"repository": "test-image"},
	}
	valuesWithPrefix := map[string]interface{}{
		"image": map[string]interface{}{"prefix": "example.io/", "repository": "test-image"},
	}
	assert.Equal(t, image.DefaultDockerRegistry+"/test-image", GetImageRefFromImageTrigger(trigger, valuesWithoutPrefix).WithoutTag())
	assert.Equal(t, "example.io/test-image", GetImageRefFromImageTrigger(trigger, valuesWithPrefix).WithoutTag())
}

func TestGenerateTemplateCommands(t *testing.T) {
	chartRef := "stable/cert-manager"
	chartVer := "v0.5.1"
	valuesFile := "values-certmanager.yaml"
	releaseFile := path.Join(os.TempDir(), "releases.yaml")
	expectedValuesFile := path.Join(os.TempDir(), valuesFile)
	releaseName := "cert-manager"
	envName := "kube-system"
	envNamespace := "kube-system"
	release := &model.Release{
		Name: releaseName,
		Chart: &model.Chart{
			Reference: &chartRef,
			Version:   &chartVer,
		},
		ValuesFile: &valuesFile,
		Triggers: []model.ReleaseUpdateTrigger{
			{Chart: &model.HelmTrigger{Track: semver.TrackMinorVersion}},
		},
		FromFile: releaseFile,
	}
	env := &model.Environment{
		Name:          envName,
		KubeNamespace: envNamespace,
	}

	tmpDir := path.Join(os.TempDir(), fmt.Sprintf("kcd-template.%d", os.Getpid()))

	t.Run("generate template commands", func(t *testing.T) {
		cmds, err := GenerateTemplateCommands(release, env)
		assert.NoError(t, err)
		assert.Equal(t, [][]string{{"mkdir", "-m", "700", "-p", tmpDir}, {"helm", "fetch", chartRef, "--version", chartVer, "--untar", "--untardir", tmpDir}, {"helm", "--kube-context", "env:" + envName, "template", tmpDir + "/" + releaseName, "-n", releaseName, "--namespace", envNamespace, "--values", expectedValuesFile}, {"rm", "-rf", tmpDir}}, cmds)
	})
}

func TestGenerateHelmDiffArgv(t *testing.T) {
	chartRef := "stable/cert-manager"
	chartVer := "v0.5.1"
	valuesFile := "values-certmanager.yaml"
	releaseFile := path.Join(os.TempDir(), "releases.yaml")
	expectedValuesFile := path.Join(os.TempDir(), valuesFile)
	releaseName := "cert-manager"
	envName := "kube-system"
	envNamespace := "kube-system"
	release := &model.Release{
		Name: releaseName,
		Chart: &model.Chart{
			Reference: &chartRef,
			Version:   &chartVer,
		},
		ValuesFile: &valuesFile,
		Triggers: []model.ReleaseUpdateTrigger{
			{Chart: &model.HelmTrigger{Track: semver.TrackMinorVersion}},
		},
		FromFile: releaseFile,
	}
	env := &model.Environment{
		Name:          envName,
		KubeNamespace: envNamespace,
	}

	t.Run("generate helm diff argv", func(t *testing.T) {
		cmds, err := GenerateHelmDiffArgv(release, env)
		assert.NoError(t, err)
		assert.Equal(t,
			[]string{
				"helm", "--kube-context", "env:" + envName, "diff", "upgrade", releaseName,
				chartRef, "--version", chartVer,
				"--values", expectedValuesFile},
			cmds)
	})

	t.Run("with env default values", func(t *testing.T) {
		env.DefaultValues = []model.ChartValue{{Key: "foo", Value: "bar"}}
		cmds, err := GenerateHelmDiffArgv(release, env)
		assert.NoError(t, err)
		assert.Equal(t,
			[]string{
				"helm", "--kube-context", "env:" + envName, "diff", "upgrade", releaseName,
				chartRef, "--version", chartVer,
				"--set-string", "foo=bar", "--values", expectedValuesFile},
			cmds)
		env.DefaultValues = nil
	})
}

func TestKubectlDiffCommand(t *testing.T) {
	t.Run("single file", func(t *testing.T) {
		files := []string{"deployment.yaml"}
		cmd := KubectlDiffCommand(files, "test-env")
		assert.Equal(t, []string{"kubectl", "--context", "env:test-env", "diff", "-f", "deployment.yaml"}, cmd)
	})

	t.Run("multiple files", func(t *testing.T) {
		files := []string{"deployment.yaml", "service.yaml"}
		cmd := KubectlDiffCommand(files, "test-env")
		assert.Equal(t, []string{"kubectl", "--context", "env:test-env", "diff", "-f", "deployment.yaml", "-f", "service.yaml"}, cmd)
	})
}

func TestDiffCommands(t *testing.T) {
	chartRef := "stable/prometheus"
	chartVer := "v1.0.0"
	releaseFile := path.Join(os.TempDir(), "releases.yaml")
	envName := "test-env"
	envNamespace := "default"

	t.Run("helm chart release", func(t *testing.T) {
		release := &model.Release{
			Name: "prometheus",
			Chart: &model.Chart{
				Reference: &chartRef,
				Version:   &chartVer,
			},
			FromFile: releaseFile,
		}
		env := &model.Environment{
			Name:          envName,
			KubeNamespace: envNamespace,
			Releases:      []*model.Release{release},
		}
		release.Environment = env

		cmds, err := DiffCommands(env, nil)
		assert.NoError(t, err)
		assert.Len(t, cmds, 1)
		assert.Equal(t,
			[]string{"helm", "--kube-context", "env:" + envName, "diff", "upgrade", "prometheus", chartRef, "--version", chartVer},
			cmds[0])
	})

	t.Run("kubectl resource files", func(t *testing.T) {
		release := &model.Release{
			Name:          "echoserver",
			ResourceFiles: []string{"echo-app.yaml"},
			FromFile:      releaseFile,
		}
		env := &model.Environment{
			Name:          envName,
			KubeNamespace: envNamespace,
			Releases:      []*model.Release{release},
		}
		release.Environment = env

		cmds, err := DiffCommands(env, nil)
		assert.NoError(t, err)
		assert.Len(t, cmds, 1)
		expectedFile := path.Join(os.TempDir(), "echo-app.yaml")
		assert.Equal(t,
			[]string{"kubectl", "--context", "env:" + envName, "diff", "-f", expectedFile},
			cmds[0])
	})

	t.Run("limit to specific releases", func(t *testing.T) {
		release1 := &model.Release{
			Name: "prometheus",
			Chart: &model.Chart{
				Reference: &chartRef,
				Version:   &chartVer,
			},
			FromFile: releaseFile,
		}
		release2 := &model.Release{
			Name:          "echoserver",
			ResourceFiles: []string{"echo-app.yaml"},
			FromFile:      releaseFile,
		}
		env := &model.Environment{
			Name:          envName,
			KubeNamespace: envNamespace,
			Releases:      []*model.Release{release1, release2},
		}
		release1.Environment = env
		release2.Environment = env

		cmds, err := DiffCommands(env, []string{"prometheus"})
		assert.NoError(t, err)
		assert.Len(t, cmds, 1)
		assert.Equal(t,
			[]string{"helm", "--kube-context", "env:" + envName, "diff", "upgrade", "prometheus", chartRef, "--version", chartVer},
			cmds[0])
	})

	t.Run("unknown release returns error", func(t *testing.T) {
		env := &model.Environment{
			Name:          envName,
			KubeNamespace: envNamespace,
			Releases:      []*model.Release{},
		}

		_, err := DiffCommands(env, []string{"unknown-release"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "release not found")
	})
}
