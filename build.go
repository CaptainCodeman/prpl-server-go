package prpl

import (
	"fmt"
	"sort"

	"path/filepath"

	"github.com/google/http2preload"
)

type (
	build struct {
		configOrder  int
		requirements capability
		entrypoint   string
		pushManifest http2preload.Manifest
	}

	builds []*build
)

func loadBuilds(config *ProjectConfig, root string) builds {
	builds := builds{}
	entrypoint := "index.html"
	if config != nil && config.Entrypoint != "" {
		entrypoint = config.Entrypoint
	}

	if config == nil || len(config.Builds) == 0 {
		fmt.Printf("WARNING: No builds configured\n")
		builds = append(builds, newBuild(config, 0, 0, entrypoint, root, root))
	} else {
		for i, build := range config.Builds {
			if build.Name == "" {
				fmt.Printf("WARNING: Build at offset %d has no name; skipping.\n", i)
				continue
			}
			builds = append(builds, newBuild(config, i, newCapabilities(build.BrowserCapabilities), "/"+filepath.Join(build.Name, entrypoint), filepath.Join(root, build.Name), root))
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
		fmt.Printf("WARNING: All builds have a capability requirement. Some browsers will display an error. Consider a fallback build.\n")
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

func newBuild(config *ProjectConfig, configOrder int, requirements capability, entrypoint, buildDir, serverRoot string) *build {
	pushManifestPath := filepath.Join(buildDir, "push-manifest.json")
	pushManifest, err := http2preload.ReadManifest(pushManifestPath)
	if err != nil {

	}

	// update paths to account for build folder
	manifest := http2preload.Manifest{}
	for path, assets := range pushManifest {
		adjusted := make(map[string]http2preload.AssetOpt, len(assets))
		for assetPath, asset := range assets {
			adjusted["/"+buildDir+"/"+assetPath] = asset
		}
		manifest["/"+buildDir+"/"+path] = adjusted
	}

	// Tidy this and verify that it's correct reasoning
	// also, knowledge of webcomponentsjs is a little too specific
	// but we could scan the additonalDependencies to find it (?)

	// add entries to push the webcomponents-loader and shell
	manifest[entrypoint] = map[string]http2preload.AssetOpt{
		"/" + buildDir + "/bower_components/webcomponentsjs/webcomponents-loader.js": {
			Type:   "script",
			Weight: 1,
		},
		"/" + buildDir + "/" + config.Shell: {
			Type:   "document",
			Weight: 1,
		},
	}

	// add entries for the routes to load the fragments
	// NOTE: These really need to be added to the router
	// as they are an exact-match so would fail as prefix
	for k, v := range config.Routes {
		manifest[k] = map[string]http2preload.AssetOpt{
			"/" + buildDir + "/" + v: {
				Type:   "document",
				Weight: 1,
			},
		}
	}

	// TODO: add child dependencies to each parent (?)

	build := build{
		configOrder:  configOrder,
		requirements: requirements,
		entrypoint:   entrypoint,
		pushManifest: manifest,
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
