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

package main

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

const yamlFixture1 = `# comment
releases:
  - name: prod-demo
    values:
      - key: fullnameOverride
        value: "demo-app"
      - key: image.tag
        value: "v1.0"
`

func TestWriteIndentedYamlToFile(t *testing.T) {
	var doc yaml.Node
	assert.NoError(t, yaml.Unmarshal([]byte(yamlFixture1), &doc))
	releases := yamlNodeMapEntry(doc.Content[0], "releases")
	for _, release := range releases.Content {
		name := yamlNodeMapEntry(release, "name")
		if name == nil || name.Value != "prod-demo" {
			continue
		}
		values := yamlNodeMapEntry(release, "values")
		if values == nil {
			continue
		}
		for _, chartValue := range values.Content {
			key := yamlNodeMapEntry(chartValue, "key")
			value := yamlNodeMapEntry(chartValue, "value")
			if key == nil || value == nil {
				continue
			}
			if key.Value == "image.tag" {
				value.Value = "v1.1"
			}
		}
	}
	expected := `# comment
releases:
  - name: prod-demo
    values:
      - key: fullnameOverride
        value: "demo-app"
      - key: image.tag
        value: "v1.1"
`
	buf, err := yaml.Marshal(&doc)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(buf))
}
