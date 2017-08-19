package prpl

import (
	"bytes"
	"time"

	"net/http"
)

type (
	Template interface {
		Render(w http.ResponseWriter, r *http.Request)
	}

	defaultTemplate struct {
		path    string
		data    []byte
		modTime time.Time
	}

	createTemplateFn func(path string, data []byte, modTime time.Time) Template
)

func createDefaultTemplate(path string, data []byte, modTime time.Time) Template {
	return &defaultTemplate{
		path:    path,
		data:    data,
		modTime: modTime,
	}
}

func (t *defaultTemplate) Render(w http.ResponseWriter, r *http.Request) {
	content := bytes.NewReader(t.data)
	http.ServeContent(w, r, t.path, t.modTime, content)
}
