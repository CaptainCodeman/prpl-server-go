package prpl

import (
	"os"

	"encoding/json"
)

// Manifest is the preload manifest map where each value,
// being a collection of resources to be preloaded,
// keyed by a serving URL path which requires those resources.
type Manifest map[string]map[string]AssetOpt

// AssetOpt defines a single resource options.
type AssetOpt struct {
	// Type is the resource type
	Type string `json:"type,omitempty"`

	// Weight is not used in the HTTP/2 preload spec
	// but some HTTP/2 servers, while implementing stream priorities,
	// could benefit from this manifest format as well.
	Weight uint8 `json:"weight,omitempty"`
}

// ReadManifest reads a push manifest from name file.
func ReadManifest(name string) (Manifest, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m Manifest
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}

	return m, nil
}
