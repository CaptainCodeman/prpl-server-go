package prpl

import (
	"regexp"
	"time"

	"net/http"
	"path/filepath"

	"github.com/google/http2preload"
)

// Matches URLs like "/foo/bar.png" but not "/foo.png/bar".
var hasFileExtension = regexp.MustCompile(`\.[^/]*$`)

// TODO Service worker location should be configurable.
var isServiceWorker = regexp.MustCompile(`service-worker.js$`)

// ServeHTTP provides standard library http server
// to apply the PRPL pattern loading
func (p *prpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	capabilities := p.browserCapabilities(r.UserAgent())
	build := p.builds.findBuild(capabilities)
	if build == nil {
		// you were warned !
		http.Error(w, "This browser is not supported", http.StatusInternalServerError)
		return
	}

	urlPath := r.URL.Path

	serveFilename, err := filepath.Abs(filepath.Join(string(p.root), urlPath))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if urlPath == "/" || !hasFileExtension.MatchString(urlPath) {
		serveFilename = build.entrypoint
	}

	file, err := p.root.Open(serveFilename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// A service worker may only register with a scope above its own path if
	// permitted by this header.
	// https://www.w3.org/TR/service-workers-1/#service-worker-allowed
	if isServiceWorker.MatchString(urlPath) {
		w.Header().Set("Service-Worker-Allowed", "/")
	}

	if build.pushManifest != nil {
		if assets, ok := build.pushManifest[urlPath]; ok {
			s := r.Header.Get("x-forwarded-proto")
			if s == "" && r.TLS != nil {
				s = "https"
			}
			if s == "" {
				s = "http"
			}
			http2preload.AddHeader(w.Header(), s, r.Host, assets)
		}
	}

	// send file
	http.ServeContent(w, r, urlPath, time.Now().UTC(), file)
}
