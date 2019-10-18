package model

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ghodss/yaml"
)

type ChartValueRef struct {
	GceResource *GceValueRef `json:"gceResource,optional"`
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
