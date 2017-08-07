package prpl

import (
	"bytes"
	"regexp"

	"net/http"

	"github.com/go-chi/chi"
)

var (
	// TODO: Service worker location should be configurable.
	// TODO: avoid doing regex at runtime, flag entry at build
	isServiceWorker = regexp.MustCompile(`service-worker.js$`)
)

func (p *prpl) createHandler() http.Handler {
	r := chi.NewRouter()

	r.Handle(p.version+"*", http.StripPrefix(p.version, http.HandlerFunc(p.staticHandler)))
	r.Get("/*", p.routeHandler)

	return r
}

func (p *prpl) routeHandler(w http.ResponseWriter, r *http.Request) {
	capabilities := p.browserCapabilities(r.UserAgent())
	build := p.builds.findBuild(capabilities)
	if build == nil {
		http.Error(w, "This browser is not supported", http.StatusInternalServerError)
		return
	}

	file, found := files[build.entrypoint]
	if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	h := w.Header()

	h.Set("Cache-Control", "public, max-age=0")
	build.addHeaders(w, h, r.URL.Path)

	content := bytes.NewReader(file.data)
	http.ServeContent(w, r, build.entrypoint, file.modTime, content)
}

func (p *prpl) staticHandler(w http.ResponseWriter, r *http.Request) {
	file, found := files[r.URL.Path]
	if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	h := w.Header()

	if isServiceWorker.MatchString(r.URL.Path) {
		h.Set("Service-Worker-Allowed", "/")
		h.Set("Cache-Control", "private, max-age=0")
	} else {
		h.Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	// TODO: if using original prpl-server-node strategy
	// add the push headers for *this* push-manifest entry
	// build.addHeaders(w, h, r.URL.Path)

	content := bytes.NewReader(file.data)
	http.ServeContent(w, r, r.URL.Path, file.modTime, content)
}

func (b *build) addHeaders(w http.ResponseWriter, header http.Header, filename string) {
	if links, ok := b.pushHeaders[filename]; ok {
		// TODO: use actual push if server supports it
		// need to add content type to push header info
		// if pusher, ok := w.(http.Pusher); ok {
		/*
			for _, url := range headers {
				pusher.Push(url, &http.PushOptions{
					Header: http.Header{
						"Cache-Control": []string{"public, max-age=31536000, immutable"},
						"Content-Type":  []string{"TODO: file content type here"},
					},
				})
			}
		*/
		// } else {
		// otherwise hope there is a proxy that will do it for us
		for _, link := range links {
			header.Add("Link", link)
		}
		// }
	}
}
