package helm

import (
	"github.com/stretchr/testify/assert"
	"github.com/kubecd/kubecd/pkg/exec"
	"github.com/kubecd/kubecd/pkg/image"
	"github.com/kubecd/kubecd/pkg/model"
	"github.com/kubecd/kubecd/pkg/semver"
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
	releaseFile := "/tmp/releases.yaml"
	expectedValuesFile := "/tmp/" + valuesFile
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
		env.DefaultValuesFile = "/tmp/env-values.yaml"
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
