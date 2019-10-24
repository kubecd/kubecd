package updates

import (
	"fmt"
	"os"

	"github.com/zedge/kubecd/pkg/helm"
	"github.com/zedge/kubecd/pkg/image"
	"github.com/zedge/kubecd/pkg/model"
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
