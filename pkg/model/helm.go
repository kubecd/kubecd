package model

import (
	"encoding/json"
	"strconv"
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
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	*is = FlexString(strconv.Itoa(i))
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
