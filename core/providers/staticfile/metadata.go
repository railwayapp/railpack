package staticfile

import (
	"path"

	"github.com/railwayapp/railpack/core/generate"
)

// Records where a static site's files are in the final image and how they
// should be served, so Railway can serve the site from its edge instead of from
// a running container.
//
// outputDir is relative to the app directory. assetsDir is the directory,
// relative to outputDir, where the framework writes files with a content hash
// in their name; it is empty when the framework has no such directory.
func SetSiteMetadata(ctx *generate.GenerateContext, outputDir, framework, assetsDir string, indexFallback bool) {
	ctx.Metadata.Set("staticOutputDir", path.Join("/app", outputDir))
	ctx.Metadata.Set("staticFramework", framework)
	ctx.Metadata.Set("staticAssetsDir", assetsDir)
	ctx.Metadata.SetBool("staticIndexFallback", indexFallback)
}
