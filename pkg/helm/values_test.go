package helm

import (
	"github.com/stretchr/testify/assert"
	"github.com/zedge/kubecd/pkg/exec"
	"github.com/zedge/kubecd/pkg/model"
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
