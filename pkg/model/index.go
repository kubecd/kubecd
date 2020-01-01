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
