package schema

import "github.com/opencontainers/go-digest"

type Descriptor struct {
	MediaType    string            `json:"mediaType"`
	Digest       digest.Digest     `json:"digest"`
	Size         int64             `json:"size"`
	Urls         []string          `json:"urls,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Data         string            `json:"data,omitempty"`
	ArtifactType string            `json:"artifactType,omitempty"`
}

type Manifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType"`
	ArtifactType  string            `json:"artifactType,omitempty"`
	Config        Descriptor        `json:"config"`
	Layers        []Descriptor      `json:"layers"`
	Subject       Descriptor        `json:"subject,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}
