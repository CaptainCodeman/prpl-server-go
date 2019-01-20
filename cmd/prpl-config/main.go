package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"path/filepath"
	"text/template"

	"github.com/captaincodeman/prpl-server-go"
)

type Context struct {
	*prpl.ProjectConfig
	ProjectID      string
	ProjectService string
	ProjectVersion string
	StaticPath     string
}

var (
	help           bool
	version        bool
	root           string
	config         string
	projectID      string
	projectService string
	projectVersion string
	staticPath     string
	templatePath   string
	yamlTemplate   *template.Template
)

func init() {
	timestamp := time.Now().Format("20060102150405")

	flag.BoolVar(&help, "help", false, "Print this help text.")
	flag.BoolVar(&version, "version", false, "Print the installed version.")
	flag.StringVar(&projectID, "project-id", "", `AppEngine Project ID.`)
	flag.StringVar(&projectService, "project-service", "default", `AppEngine Service.`)
	flag.StringVar(&projectVersion, "project-version", "", `AppEngine Version.`)
	flag.StringVar(&staticPath, "static-path", "static", `Path to static folder relative to app.yaml (default "static").`)
	flag.StringVar(&root, "root", ".", `Serve files relative to this directory (default ".").`)
	flag.StringVar(&config, "config", "", `JSON configuration file (default "<root>/polymer.json" if exists).`)
	flag.StringVar(&templatePath, "template", "", `app.yaml template file (default inbuilt template).`)
}

func main() {
	flag.Parse()

	if help {
		flag.PrintDefaults()
		return
	}

	if version {
		fmt.Printf("Version 0.0.1/n")
		return
	}

	if root == "" {
		fmt.Printf("invalid --root")
		return
	}

	if config == "" {
		config = "polymer.json"
	}

	filename := filepath.Join(root, config)
	config, err := prpl.ConfigFromFile(filename)
	if err != nil {
		fmt.Printf("couldn't load config %v\n", err)
		return
	}

	if templatePath == "" {
		yamlTemplate = template.Must(template.New("template").Parse(templateString))
	} else {
		var err error
		yamlTemplate, err = template.ParseFiles(templatePath)
		if err != nil {
			fmt.Printf("couldn't load template %v\n", err)
			return
		}
	}

	context := Context{
		config,
		projectID,
		projectService,
		projectVersion,
		staticPath,
	}

	if err := yamlTemplate.Execute(os.Stdout, context); err != nil {
		fmt.Printf("error %v\n", err)
	}
}

const templateString = `{{ if .ProjectID }}project: {{ .ProjectID }}{{ end }}
{{ if .ProjectVersion }}version: {{ .ProjectVersion }}{{ end }}
service: {{ .ProjectService }}
runtime: go
api_version: go1.8

instance_class: F1

handlers:
{{- with $x := . -}}
{{- range $build := .Builds }}
- url: /{{ $build.Name }}/index.html
  script: _go_app
  secure: always

- url: /{{ $build.Name }}/service-worker.js
  static_files: {{ $x.StaticPath }}/{{ $build.Name }}/service-worker.js
  upload: {{ $x.StaticPath }}/{{ $build.Name }}/service-worker.js
  secure: always
  http_headers:
    Cache-Control: "private, max-age=0, must-revalidate"
    Service-Worker-Allowed: "/"

- url: /{{ $build.Name }}/
  static_dir: {{ $x.StaticPath }}/{{ $build.Name }}/
  application_readable: true
  secure: always
  http_headers:
    Cache-Control: "public, max-age=31536000, immutable"
{{ end }}
{{ end -}}

- url: /.*
  script: _go_app
  secure: always
`
