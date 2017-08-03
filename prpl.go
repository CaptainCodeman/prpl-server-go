package prpl

import (
	"net/http"

	"github.com/ua-parser/uap-go/uaparser"
)

type (
	// prpl is an instance of the prpl-server service
	prpl struct {
		parser *uaparser.Parser
		config *ProjectConfig
		builds builds
		root   http.Dir
	}

	// optionFn provides functonal option configuration
	optionFn func(*prpl) error
)

// New creates a new prpl instance
func New(options ...optionFn) (*prpl, error) {
	p := prpl{
		parser: uaparser.NewFromSaved(),
		root:   http.Dir("."),
	}

	for _, option := range options {
		if err := option(&p); err != nil {
			return nil, err
		}
	}

	// use polymer.json for build file by default
	if p.config == nil {
		if err := ConfigFile("polymer.json")(&p); err != nil {
			return nil, err
		}
	}

	return &p, nil
}

func Root(root http.Dir) optionFn {
	return func(p *prpl) error {
		p.root = root
		return nil
	}
}

// ConfigFile loads the project configuration
func ConfigFile(filename string) optionFn {
	return func(p *prpl) error {
		config, err := loadProjectConfig(filename)
		if err != nil {
			// return err
		}
		p.config = config
		p.builds = loadBuilds(string(p.root), config)
		return nil
	}
}

// RegexFile allows the uaparser configuration
// to be overriden from the provided yaml file
func RegexFile(regexFile string) optionFn {
	return func(p *prpl) error {
		parser, err := uaparser.New(regexFile)
		if err != nil {
			return err
		}
		p.parser = parser
		return nil
	}
}

// RegexBytes allows the uaparser configuration
// to be overriden from the yaml-formatted bytes
func RegexBytes(data []byte) optionFn {
	return func(p *prpl) error {
		parser, err := uaparser.NewFromBytes(data)
		if err != nil {
			return err
		}
		p.parser = parser
		return nil
	}
}
