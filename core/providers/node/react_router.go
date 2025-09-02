package node

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DefaultReactRouterOutputDirectory = "build/client/"
	ReactRouterConfigJS               = "react-router.config.js"
	ReactRouterConfigTS               = "react-router.config.ts"
)

// Detect React Router app using:
// - presence of a react-router.config.{ts,js} file (optionally scoped to a workspace path)
// - a build script containing "react-router build"
// - having a dependency on "@react-router/dev"
// and ensure a build script exists so there is an actual output directory
func (p *NodeProvider) isReactRouter(ctx *generate.GenerateContext) bool {
	pkg := p.workspace.Root
	if pkg == nil || pkg.PackageJson == nil {
		return false
	}

	rrConfigJS := ReactRouterConfigJS
	rrConfigTS := ReactRouterConfigTS
	if pkg.Path != "" {
		rrConfigJS = pkg.Path + "/" + ReactRouterConfigJS
		rrConfigTS = pkg.Path + "/" + ReactRouterConfigTS
	}

	hasRRConfig := ctx.App.HasMatch(rrConfigJS) || ctx.App.HasMatch(rrConfigTS)
	hasBuildCommand := pkg.PackageJson.HasScript("build")
	hasRRBuildCommand := strings.Contains(strings.ToLower(pkg.PackageJson.GetScript("build")), "react-router build")
	// TODO we could probably just check for this? Need to understand how older versions of react-router work more.
	hasRRPackage := pkg.PackageJson.hasDependency("@react-router/dev")

	return hasBuildCommand && (hasRRConfig || hasRRBuildCommand || hasRRPackage)
}

// getReactRouterOutputDirectory attempts to read the output directory from react-router.config.{ts,js} by extracting
// the buildDirectory option in the config object.
func (p *NodeProvider) getReactRouterOutputDirectory(ctx *generate.GenerateContext) string {
	configContent := ""

	if ctx.App.HasMatch(ReactRouterConfigJS) {
		if content, err := ctx.App.ReadFile(ReactRouterConfigJS); err == nil {
			configContent = content
		}
	} else if ctx.App.HasMatch(ReactRouterConfigTS) {
		if content, err := ctx.App.ReadFile(ReactRouterConfigTS); err == nil {
			configContent = content
		}
	}

	if configContent != "" {
		// TODO this field can be an expression `buildDirectory: "build/" + process.env.NODE_ENV,` so we should tighten
		// up the regex here (and in the vite provider, since vite config can have expressions too)
		// `buildDirectory: "custom-directory/"`
		buildDirRegex := regexp.MustCompile(`buildDirectory:\s*['"](.+?)['"]`)
		if matches := buildDirRegex.FindStringSubmatch(configContent); len(matches) > 1 {
			return matches[1]
		}
	}

	return DefaultReactRouterOutputDirectory
}
