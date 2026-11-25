package staticfile

import (
	"path"

	"github.com/railwayapp/railpack/core/generate"
)

// Caddy runs from the app directory, so a relative root is relative to it.
const appDir = "/app"

// The absolute directory Caddy serves for a root given as Railpack config.
func servedDir(root string) string {
	if path.IsAbs(root) {
		return path.Clean(root)
	}
	return path.Join(appDir, root)
}

// Records where a static site's files are in the final image and how they
// should be served, so Railway can serve the site from its edge instead of from
// a running container.
//
// outputDir is the absolute path Caddy serves from. assetsDir is the
// directory, relative to outputDir, where the framework writes files with a
// content hash in their name; it is empty when the framework has no such
// directory.
func SetSiteMetadata(ctx *generate.GenerateContext, outputDir, framework, assetsDir string, indexFallback bool) {
	ctx.Metadata.Set("staticOutputDir", path.Clean(outputDir))
	ctx.Metadata.Set("staticFramework", framework)
	ctx.Metadata.Set("staticAssetsDir", assetsDir)
	ctx.Metadata.SetBool("staticIndexFallback", indexFallback)
}
