package node

import (
	"path"

	"github.com/railwayapp/railpack/core/generate"
)

var turboConfigFiles = []string{"turbo.json", "turbo.jsonc"}

// turboCacheDir is where turbo keeps its local task cache, or "" when the app does not use turbo.
// Kept across builds as a cache mount, it lets unchanged workspace packages replay their outputs
// instead of rebuilding: TURBO_CACHE_DIR, then the config's cacheDir, then turbo's .turbo/cache,
// relative to the repository root.
func (p *NodeProvider) turboCacheDir(ctx *generate.GenerateContext) string {
	config, _, _ := ctx.App.ReadFirstFileOf(turboConfigFiles...)
	if config == "" && (p.packageJson == nil || !p.packageJson.hasDependency("turbo")) {
		return ""
	}

	dir := ".turbo/cache"
	var turbo struct {
		CacheDir string `json:"cacheDir"`
	}
	if config != "" && ctx.App.ReadJSON(config, &turbo) == nil && turbo.CacheDir != "" {
		dir = turbo.CacheDir
	}
	if v := ctx.Env.GetVariable("TURBO_CACHE_DIR"); v != "" {
		dir = v
	}
	// Cache mount targets use Linux paths: preserve absolute overrides and normalize
	// redundant separators or dot segments so the mount matches turbo's cache directory.
	if path.IsAbs(dir) {
		return path.Clean(dir)
	}

	return path.Join("/app", dir)
}
