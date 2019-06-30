package model

type ChartValueRef struct {
	GceResource *GceValueRef `json:"gceResource,optional"`
}

type ChartValue struct {
	Key       string         `json:"key"`
	Value     string         `json:"value,omitempty"`
	ValueFrom *ChartValueRef `json:"valueFrom,omitempty"`
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
