package prpl

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"io/ioutil"
	"net/http"
	"path/filepath"
)

type (
	build struct {
		name         string
		configOrder  int
		requirements capability
		entrypoint   string
		template     Template
		pushHeaders  PushHeaders
	}

	builds []*build

	file struct {
		data    []byte
		size    int64
		modTime time.Time
	}

	// TODO: add all headers including cache-control and
	// service worker so no regex matching is needed at runtime

	// PushHeaders are the link headers to send for a route
	PushHeaders map[string][]string
)

var files = make(map[string]*file)

func loadBuilds(config *ProjectConfig, root http.Dir, routes Routes, version string, createTemplate createTemplateFn) builds {
	builds := builds{}
	entrypoint := "index.html"
	if config != nil && config.Entrypoint != "" {
		entrypoint = config.Entrypoint
	}

	if config == nil || len(config.Builds) == 0 {
		log.Println("WARNING: No builds configured")
		builds = append(builds, newBuild(config, 0, "", 0, entrypoint, string(root), root, routes, version, createTemplate))
	} else {
		for i, build := range config.Builds {
			if build.Name == "" {
				log.Printf("WARNING: Build at offset %d has no name; skipping.\n", i)
				continue
			}
			builds = append(builds, newBuild(config, i, build.Name, newCapabilities(build.BrowserCapabilities), filepath.Join(build.Name, entrypoint), filepath.Join(string(root), build.Name), root, routes, version, createTemplate))
		}
	}

	sort.Sort(byPriority(builds))

	// Sanity check.
	fallbackFound := false
	for _, build := range builds {

		// Note `build.entrypoint` is relative to the server root, but that's not
		// neccessarily our cwd.
		// TODO Refactor to make filepath vs URL path and relative vs absolute
		// values clearer.
		// if (!fs.existsSync(path.join(root, build.entrypoint))) {
		//   console.warn(`WARNING: Entrypoint "${build.entrypoint}" does not exist.`);
		// }

		if build.requirements == 0 {
			fallbackFound = true
		}
	}

	if !fallbackFound {
		log.Println("WARNING: All builds have a capability requirement. Some browsers will display an error. Consider a fallback build.")
	}

	return builds
}

type byPriority builds

func (a byPriority) Len() int      { return len(a) }
func (a byPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byPriority) Less(i, j int) bool {
	sizeDiff := a[i].requirements.size() - a[j].requirements.size()
	if sizeDiff == 0 {
		return a[i].configOrder < a[j].configOrder
	}
	return sizeDiff > 0
}

func newBuild(config *ProjectConfig, configOrder int, name string, requirements capability, entrypoint, buildDir string, root http.Dir, routes Routes, version string, createTemplate createTemplateFn) *build {
	pushManifestPath := filepath.Join(buildDir, "push-manifest.json")
	pushManifest, err := ReadManifest(pushManifestPath)
	if err != nil {
		// return err
	}

	var template Template

	err = filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		file := &file{
			size:    info.Size(),
			modTime: info.ModTime(),
		}

		filename, _ := filepath.Rel(string(root), path)
		if filename == entrypoint {
			f, err := root.Open(filename)
			if err != nil {
				return err
			}

			data, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			// add version to path
			data = bytes.Replace(
				data,
				[]byte(fmt.Sprintf(`<base href="/%s/">`, name)),
				[]byte(fmt.Sprintf(`<base href="%s%s/">`, version, name)),
				1)

			template = createTemplate(entrypoint, data, info.ModTime())

			file.data = data
			files[filename] = file
		}

		return nil
	})
	if err != nil {
		// return err
	}

	// create map of routes -> push headers
	pushHeaders := PushHeaders{}
	prefix := version + name + "/"

	for path, fragment := range routes {
		set := map[string]struct{}{}
		headers := []string{
			fmt.Sprintf("<%s%s>; rel=preload; as=%s", prefix, "bower_components/webcomponentsjs/webcomponents-loader.js", "script"),
			fmt.Sprintf("<%s%s>; rel=preload; as=%s", prefix, config.Shell, "document"),
		}
		set[headers[0]] = struct{}{}
		set[headers[1]] = struct{}{}
		for path, asset := range pushManifest[config.Shell] {
			link := fmt.Sprintf("<%s%s>; rel=preload; as=%s", prefix, path, asset.Type)
			if _, found := set[link]; !found {
				set[link] = struct{}{}
				headers = append(headers, link)
			}
		}

		headers = append(headers, fmt.Sprintf("<%s%s>; rel=preload; as=%s", prefix, fragment, "document"))
		for path, asset := range pushManifest[fragment] {
			link := fmt.Sprintf("<%s%s>; rel=preload; as=%s", prefix, path, asset.Type)
			if _, found := set[link]; !found {
				set[link] = struct{}{}
				headers = append(headers, link)
			}
		}

		pushHeaders[path] = headers
	}

	// update paths to account for the build folder name
	manifest := Manifest{}
	for path, assets := range pushManifest {
		adjusted := make(map[string]AssetOpt, len(assets))
		for assetPath, asset := range assets {
			adjusted[prefix+assetPath] = asset
		}
		manifest[prefix+path] = adjusted
	}

	build := build{
		name:         name,
		configOrder:  configOrder,
		requirements: requirements,
		entrypoint:   entrypoint,
		template:     template,
		pushHeaders:  pushHeaders,
	}

	return &build
}

func (b *build) canServe(client capability) bool {
	return client&b.requirements == b.requirements
}

func (b builds) findBuild(client capability) *build {
	for _, build := range b {
		if build.canServe(client) {
			return build
		}
	}

	return nil
}
