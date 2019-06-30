package model

type ImageTrigger struct {
	TagValue        string `json:"tagValue"`        // TODO: default to "image.tag"
	RepoValue       string `json:"repoValue"`       // TODO: default to "image.repository"
	RepoPrefixValue string `json:"repoPrefixValue"` // TODO: default to "image.prefix"
	Track           string `json:"track"`           // one of "PatchLevel", "MinorVersion", "MajorVersion", "Newest"
}

type HelmTrigger struct {
	Track string `json:"track"` // one of "PatchLevel", "MinorVersion", "MajorVersion", "Newest"
}

type ReleaseUpdateTrigger struct {
	Image *ImageTrigger `json:"image,omitempty"`
	Chart *HelmTrigger  `json:"chart,omitempty"`
}
