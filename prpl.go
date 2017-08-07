package prpl

import (
	"net/http"

	"github.com/ua-parser/uap-go/uaparser"
)

type (
	// prpl is an instance of the prpl-server service
	prpl struct {
		http.Handler
		parser  *uaparser.Parser
		config  *ProjectConfig
		builds  builds
		root    http.Dir
		routes  Routes
		version string
	}

	// optionFn provides functonal option configuration
	optionFn func(*prpl) error
)

// New creates a new prpl instance
func New(options ...optionFn) (*prpl, error) {
	p := prpl{
		parser:  uaparser.NewFromSaved(),
		root:    http.Dir("."),
		version: "/static/",
	}

	for _, option := range options {
		if err := option(&p); err != nil {
			return nil, err
		}
	}

	// use polymer.json for build file by default
	if p.config == nil {
		if err := WithConfigFile("polymer.json")(&p); err != nil {
			return nil, err
		}
	}

	p.Handler = p.createHandler()
	p.builds = loadBuilds(p.config, p.root, p.routes, p.version)

	return &p, nil
}

// TODO: provide options to auto-create the version
// based on last modified timestamp or content hash

// WithVersion sets the version prefix
func WithVersion(version string) optionFn {
	return func(p *prpl) error {
		p.version = "/" + version + "/"
		return nil
	}
}

// WithRoutes sets the route -> fragment mapping
func WithRoutes(routes Routes) optionFn {
	return func(p *prpl) error {
		p.routes = routes
		return nil
	}
}

// WithRoot sets the root directory
func WithRoot(root http.Dir) optionFn {
	return func(p *prpl) error {
		p.root = root
		return nil
	}
}

// WithConfig sets the project configuration
func WithConfig(config *ProjectConfig) optionFn {
	return func(p *prpl) error {
		p.config = config
		return nil
	}
}

// WithConfigFile loads the project configuration
func WithConfigFile(filename string) optionFn {
	return func(p *prpl) error {
		config, err := ConfigFromFile(filename)
		if err != nil {
			return err
		}
		p.config = config
		return nil
	}
}

// WithUAParserFile allows the uaparser configuration
// to be overriden from the inbuilt settings
func WithUAParserFile(regexFile string) optionFn {
	return func(p *prpl) error {
		parser, err := uaparser.New(regexFile)
		if err != nil {
			return err
		}
		p.parser = parser
		return nil
	}
}

// WithUAParserBytes allows the uaparser configuration
// to be overriden from the inbuilt settings
func WithUAParserBytes(data []byte) optionFn {
	return func(p *prpl) error {
		parser, err := uaparser.NewFromBytes(data)
		if err != nil {
			return err
		}
		p.parser = parser
		return nil
	}
}
