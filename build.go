package prpl

import (
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

func loadBuilds(root string, config *ProjectConfig) builds {
	builds := builds{}
	entrypoint := "index.html"
	if config != nil && config.Entrypoint != "" {
		entrypoint = config.Entrypoint
	}

	if config == nil || len(config.Builds) == 0 {
		// WARNING: No builds configure
		builds = append(builds, newBuild(0, 0, entrypoint, root, root))
	} else {
		for i, build := range config.Builds {
			if build.Name == "" {
				// WARNING: Build at offset ${i} has no name; skipping.
				continue
			}
			builds = append(builds, newBuild(i, newCapabilities(build.BrowserCapabilities), filepath.Join(build.Name, entrypoint), filepath.Join(root, build.Name), root))
		}
	}

	sort.Sort(byPriority(builds))

	// Sanity check.
	fallbackFound := false
	for _, build := range builds {
		// TODO: log
		// fmt.Sprintf(`Registered entrypoint "%s" with capabilities %s`, build.entrypoint, build.requirements)

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
		// WARNING: All builds have a capability requirement. Some browsers will display an error. Consider a fallback build.
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

func newBuild(configOrder int, requirements capability, entrypoint, buildDir, serverRoot string) *build {
	pushManifestPath := filepath.Join(buildDir, "push-manifest.json")
	pushManifest, err := http2preload.ReadManifest(pushManifestPath)
	if err != nil {

	}

	build := build{
		configOrder:  configOrder,
		requirements: requirements,
		entrypoint:   entrypoint,
		pushManifest: pushManifest,
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
