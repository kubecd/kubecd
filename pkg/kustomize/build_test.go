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

package kustomize

import (
	"testing"

	"github.com/kubecd/kubecd/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerateBuildCommand(t *testing.T) {
	t.Run("valid release", func(t *testing.T) {
		rel := &model.Release{
			Name:          "my-app",
			Kustomization: &model.Kustomization{Path: "./overlays/prod"},
			FromFile:      "/path/to/releases.yaml",
		}

		cmd := GenerateBuildCommand(rel)

		assert.Equal(t, []string{"kustomize", "build", "/path/to/overlays/prod"}, cmd)
	})

	t.Run("nil release", func(t *testing.T) {
		cmd := GenerateBuildCommand(nil)
		assert.Nil(t, cmd)
	})

	t.Run("release without kustomization", func(t *testing.T) {
		rel := &model.Release{
			Name:     "my-app",
			FromFile: "/path/to/releases.yaml",
		}
		cmd := GenerateBuildCommand(rel)
		assert.Nil(t, cmd)
	})
}

func TestGenerateApplyCommand(t *testing.T) {
	rel := &model.Release{
		Name:          "my-app",
		Kustomization: &model.Kustomization{Path: "./overlays/prod"},
		FromFile:      "/path/to/releases.yaml",
	}
	env := &model.Environment{
		Name:          "production",
		KubeNamespace: "default",
	}

	t.Run("without dry-run", func(t *testing.T) {
		cmd := GenerateApplyCommand(rel, env, false)
		expected := []string{
			"kubectl", "--context", "env:production",
			"apply", "-k", "/path/to/overlays/prod",
			"--namespace", "default",
		}
		assert.Equal(t, expected, cmd)
	})

	t.Run("with dry-run", func(t *testing.T) {
		cmd := GenerateApplyCommand(rel, env, true)
		expected := []string{
			"kubectl", "--context", "env:production",
			"apply", "-k", "/path/to/overlays/prod",
			"--dry-run=client",
			"--namespace", "default",
		}
		assert.Equal(t, expected, cmd)
	})

	t.Run("nil release", func(t *testing.T) {
		cmd := GenerateApplyCommand(nil, env, false)
		assert.Nil(t, cmd)
	})

	t.Run("release without kustomization", func(t *testing.T) {
		relNoKust := &model.Release{
			Name:     "my-app",
			FromFile: "/path/to/releases.yaml",
		}
		cmd := GenerateApplyCommand(relNoKust, env, false)
		assert.Nil(t, cmd)
	})
}

func TestGenerateTemplateCommands(t *testing.T) {
	rel := &model.Release{
		Name:          "my-app",
		Kustomization: &model.Kustomization{Path: "./overlays/prod"},
		FromFile:      "/path/to/releases.yaml",
	}
	env := &model.Environment{
		Name:          "production",
		KubeNamespace: "default",
	}

	t.Run("valid release", func(t *testing.T) {
		cmds := GenerateTemplateCommands(rel, env)

		assert.Len(t, cmds, 4)
		assert.Equal(t, []string{"echo", "---"}, cmds[0])
		assert.Equal(t, []string{"echo", "#", "Kustomization:", "/path/to/overlays/prod"}, cmds[1])
		assert.Equal(t, []string{"kustomize", "build", "/path/to/overlays/prod"}, cmds[2])
		assert.Equal(t, []string{"echo", "---"}, cmds[3])
	})

	t.Run("nil release", func(t *testing.T) {
		cmds := GenerateTemplateCommands(nil, env)
		assert.Nil(t, cmds)
	})

	t.Run("release without kustomization", func(t *testing.T) {
		relNoKust := &model.Release{
			Name:     "my-app",
			FromFile: "/path/to/releases.yaml",
		}
		cmds := GenerateTemplateCommands(relNoKust, env)
		assert.Nil(t, cmds)
	})
}
