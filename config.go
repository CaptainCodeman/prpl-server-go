package prpl

import (
	"bytes"
	"io"
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
		Shell      string        `json:"shell"`
		Builds     []BuildConfig `json:"builds"`
	}

	// BuildConfig contains the build-specific browser capabilities
	BuildConfig struct {
		Name                string   `json:"name"`
		BrowserCapabilities []string `json:"browserCapabilities"`
	}

	// Routes map urls to fragments
	Routes map[string]string
)

func ConfigFromFile(filename string) (*ProjectConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ConfigFromReader(file)
}

func ConfigFromBytes(b []byte) (*ProjectConfig, error) {
	r := bytes.NewReader(b)
	return ConfigFromReader(r)
}

func ConfigFromReader(r io.Reader) (*ProjectConfig, error) {
	var config ProjectConfig
	dec := json.NewDecoder(r)
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
