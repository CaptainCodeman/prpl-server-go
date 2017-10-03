package prpl

import (
	"bytes"
	"strings"

	"net/http"
)

func (p *prpl) createHandler() http.Handler {
	m := http.NewServeMux()

	for _, build := range p.builds {
		m.HandleFunc(p.version+build.entrypoint, p.routeHandler)
		for path, handler := range p.staticHandlers {
			m.Handle(p.version+build.name+"/"+path, handler)
		}
	}

	m.Handle(p.version, http.StripPrefix(p.version, p.staticHandler(http.FileServer(p.root))))

	m.HandleFunc("/", p.routeHandler)

	return m
}

func (p *prpl) routeHandler(w http.ResponseWriter, r *http.Request) {
	capabilities := p.browserCapabilities(r.UserAgent())
	build := p.builds.findBuild(capabilities)
	if build == nil {
		http.Error(w, "This browser is not supported", http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Cache-Control", "public, max-age=0")
	build.addHeaders(w, h, r.URL.Path)
	build.template.Render(w, r)
}

func (p *prpl) staticHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// TODO: Service worker location should be configurable.
		h := w.Header()
		if strings.HasSuffix(r.URL.Path, "service-worker.js") {
			h.Set("Service-Worker-Allowed", "/")
			h.Set("Cache-Control", "private, max-age=0")
		} else {
			h.Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		file, found := files[r.URL.Path]
		if !found {
			next.ServeHTTP(w, r)
			return
		}

		// TODO: if using original prpl-server-node strategy
		// add the push headers for *this* push-manifest entry
		// build.addHeaders(w, h, r.URL.Path)

		content := bytes.NewReader(file.data)
		http.ServeContent(w, r, r.URL.Path, file.modTime, content)
	}

	return http.HandlerFunc(fn)
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
