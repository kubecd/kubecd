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

package model

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelease_Kustomization_UnmarshalJSON(t *testing.T) {
	yamlData := []byte(`
name: kustomize-release
kustomization:
  path: ./overlays/production
`)
	var release Release
	require.NoError(t, yaml.Unmarshal(yamlData, &release))
	assert.Equal(t, "kustomize-release", release.Name)
	assert.NotNil(t, release.Kustomization)
	assert.Equal(t, "./overlays/production", release.Kustomization.Path)
	assert.Nil(t, release.Chart)
}

func TestRelease_Kustomization_SanityCheck(t *testing.T) {
	tests := []struct {
		name         string
		release      Release
		expectErrors bool
	}{
		{
			name: "valid kustomization release",
			release: Release{
				Name:          "valid-kustomize",
				Kustomization: &Kustomization{Path: "./overlays/prod"},
			},
			expectErrors: false,
		},
		{
			name: "kustomization without path",
			release: Release{
				Name:          "invalid-kustomize",
				Kustomization: &Kustomization{Path: ""},
			},
			expectErrors: true,
		},
		{
			name: "kustomization and chart both defined",
			release: Release{
				Name:          "invalid-both",
				Kustomization: &Kustomization{Path: "./overlays/prod"},
				Chart:         &Chart{Dir: strPtr("./charts/myapp")},
			},
			expectErrors: true,
		},
		{
			name: "kustomization and resourceFiles both defined",
			release: Release{
				Name:          "invalid-both",
				Kustomization: &Kustomization{Path: "./overlays/prod"},
				ResourceFiles: []string{"./resources/pod.yaml"},
			},
			expectErrors: true,
		},
		{
			name: "no chart, kustomization, or resourceFiles",
			release: Release{
				Name: "empty-release",
			},
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := tt.release.sanityCheck()
			if tt.expectErrors {
				assert.NotEmpty(t, issues, "expected errors but got none")
			} else {
				assert.Empty(t, issues, "expected no errors but got: %v", issues)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
