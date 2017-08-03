package prpl

import (
	"os"

	"encoding/json"
)

type (
	// ProjectConfig is the subset of the polymer.json
	// specification that we care about for serving.
	// https://www.polymer-project.org/2.0/docs/tools/polymer-json
	// https://github.com/Polymer/polymer-project-config/blob/master/src/index.ts
	ProjectConfig struct {
		Entrypoint string        `json:"entrypoint"`
		Builds     []BuildConfig `json:"builds"`
	}

	// BuildConfig contains the build-specific browser capabilities
	BuildConfig struct {
		Name                string   `json:"name"`
		BrowserCapabilities []string `json:"browserCapabilities"`
	}
)

func loadProjectConfig(filename string) (*ProjectConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	var config ProjectConfig
	dec := json.NewDecoder(file)
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
