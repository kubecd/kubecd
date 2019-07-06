package model

const (
	DefaultTagValue        = "image.tag"
	DefaultRepoValue       = "image.repository"
	DefaultRepoPrefixValue = "image.prefix"
)

type ImageTrigger struct {
	TagValue        string `json:"tagValue"`
	RepoValue       string `json:"repoValue"`
	RepoPrefixValue string `json:"repoPrefixValue"`
	Track           string `json:"track"` // one of "PatchLevel", "MinorVersion", "MajorVersion", "Newest"
}

func (t *ImageTrigger) TagValueString() string {
	if t.TagValue == "" {
		return DefaultTagValue
	}
	return t.TagValue
}

func (t *ImageTrigger) RepoValueString() string {
	if t.RepoValue == "" {
		return DefaultRepoValue
	}
	return t.RepoValue
}

func (t *ImageTrigger) RepoPrefixValueString() string {
	if t.RepoPrefixValue == "" {
		return DefaultRepoPrefixValue
	}
	return t.RepoPrefixValue
}

type HelmTrigger struct {
	Track string `json:"track"` // one of "PatchLevel", "MinorVersion", "MajorVersion", "Newest"
}

type ReleaseUpdateTrigger struct {
	Image *ImageTrigger `json:"image,omitempty"`
	Chart *HelmTrigger  `json:"chart,omitempty"`
}
