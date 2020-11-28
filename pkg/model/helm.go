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
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/buildkite/interpolate"
	"github.com/ghodss/yaml"
)

type ChartValueRef struct {
	GceResource *GceValueRef `json:"gceResource,omitempty"`
}

type FlexString string

type ChartValue struct {
	Key        string         `json:"key"`
	InputValue FlexString     `json:"value,omitempty"`
	Value      string         `json:"-"`
	ValueFrom  *ChartValueRef `json:"valueFrom,omitempty"`
}

type Chart struct {
	Reference *string `json:"reference,omitempty"`
	Dir       *string `json:"dir,omitempty"`
	Version   *string `json:"version,omitempty"`
}

type HelmRepo struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	CAFile   string `json:"caFile,omitempty"`
	CertFile string `json:"certFile,omitempty"`
	KeyFile  string `json:"keyFile,omitempty"`
}

func interpolateValue(val string) string {
	env := interpolate.NewSliceEnv(os.Environ())
	output, err := interpolate.Interpolate(env, val)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid interpolation: %v\n", err)
		return val
	}
	return output
}

func (r HelmRepo) GetCAFile() string {
	return interpolateValue(r.CAFile)
}

func (r HelmRepo) GetCertFile() string {
	return interpolateValue(r.CertFile)
}

func (r HelmRepo) GetKeyFile() string {
	return interpolateValue(r.KeyFile)
}

func (is *FlexString) UnmarshalJSON(data []byte) error {
	if string(data) == "true" || string(data) == "false" {
		data = []byte(`"` + string(data) + `"`)
	}
	if data[0] == '"' {
		return json.Unmarshal(data, (*string)(is))
	}
	var i int
	err := json.Unmarshal(data, &i)
	if err == nil {
		*is = FlexString(strconv.Itoa(i))
		return nil
	}
	var f float64
	err = json.Unmarshal(data, &f)
	if err == nil {
		*is = FlexString(fmt.Sprintf("%g", f))
		return nil
	}
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return err
	}

	data2, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	*is = FlexString(data2)
	return nil
}

func (v *ChartValue) UnmarshalJSON(data []byte) error {
	type chartValue ChartValue
	if err := json.Unmarshal(data, (*chartValue)(v)); err != nil {
		return err
	}
	if v.InputValue != "" {
		v.Value = string(v.InputValue)
	}
	return nil
}
