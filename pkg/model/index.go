package model

type imageIndexEntry struct {
	environment *Environment
	release     *Release
}

// imageIndex contains a map from image:tag to a list of [environment,release] tuples
// that consume the key image
type imageIndex map[string][]imageIndexEntry

//func buildImageIndex(env *Environment) imageIndex {
//	index := make(imageIndex, 0)
//	for _, release := range env.AllReleases() {
//		index[]
//	}
//	return index
//}
