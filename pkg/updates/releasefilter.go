/*
 * Copyright 2018-2020 Zedge, Inc.
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

package updates

import (
	"fmt"
	"os"

	"github.com/kubecd/kubecd/pkg/helm"
	"github.com/kubecd/kubecd/pkg/image"
	"github.com/kubecd/kubecd/pkg/model"
)

func ClusterReleaseFilter(clusterName string) ReleaseFilterFunc {
	return func(release *model.Release) bool {
		return release.Environment.Cluster.Name == clusterName
	}
}

func EnvironmentReleaseFilter(envName string) ReleaseFilterFunc {
	return func(release *model.Release) bool {
		return release.Environment.Name == envName
	}
}

func ImageReleaseFilter(imageRepo string) ReleaseFilterFunc {
	return func(release *model.Release) bool {
		imageRef := image.NewDockerImageRef(imageRepo)
		values, err := helm.GetResolvedValues(release)
		if err != nil {
			return false
		}

		for _, releaseImageRef := range helm.GetImageRefsFromRelease(release, values) {
			if releaseImageRef == nil {
				_, _ = fmt.Fprintf(os.Stderr, "WARNING: could not find image for release %q in env %q\n", release.Name, release.Environment.Name)
				return false
			}
			if releaseImageRef.WithoutTag() == imageRef.WithoutTag() {
				return true
			}
		}
		return false
	}
}

func ReleaseFilter(releaseNames []string) ReleaseFilterFunc {
	return func(release *model.Release) bool {
		for _, relName := range releaseNames {
			if relName == release.Name {
				return true
			}
		}
		return false
	}
}
