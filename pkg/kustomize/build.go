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
	"github.com/kubecd/kubecd/pkg/model"
)

// GenerateBuildCommand generates the kustomize build command for a release.
func GenerateBuildCommand(rel *model.Release) []string {
	kustomizePath := rel.AbsPath(rel.Kustomization.Path)
	return []string{"kustomize", "build", kustomizePath}
}

// GenerateApplyCommand generates the kubectl apply -k command for kustomize releases.
func GenerateApplyCommand(rel *model.Release, env *model.Environment, dryRun bool) []string {
	kustomizePath := rel.AbsPath(rel.Kustomization.Path)
	cmd := []string{"kubectl", "--context", model.KubeContextName(env.Name), "apply", "-k", kustomizePath}
	if dryRun {
		cmd = append(cmd, "--dry-run=client")
	}
	cmd = append(cmd, "--namespace", env.KubeNamespace)
	return cmd
}

// GenerateTemplateCommands generates commands to show kustomize output.
func GenerateTemplateCommands(rel *model.Release, env *model.Environment) [][]string {
	kustomizePath := rel.AbsPath(rel.Kustomization.Path)
	return [][]string{
		{"echo", "---"},
		{"echo", "#", "Kustomization:", kustomizePath},
		{"kustomize", "build", kustomizePath},
		{"echo", "---"},
	}
}
