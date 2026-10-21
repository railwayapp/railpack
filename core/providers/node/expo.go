package node

import (
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	// expo export --platform web writes the static web bundle here by default
	DefaultExpoOutputDirectory = "dist"
	DefaultExpoStartCommand    = "expo start"
)

// expoAppConfig models the subset of app.json we care about for web detection.
type expoAppConfig struct {
	Expo struct {
		Web struct {
			Output string `json:"output"`
		} `json:"web"`
	} `json:"expo"`
}

func (p *NodeProvider) isExpo(ctx *generate.GenerateContext) bool {
	return p.hasDependency("expo")
}

// isExpoSPA reports whether the Expo project is configured to export a static
// web build. Expo produces a single-page app when web output is "static" or
// "single" and react-native-web is present to render the app in the browser.
func (p *NodeProvider) isExpoSPA(ctx *generate.GenerateContext) bool {
	if !p.isExpo(ctx) {
		return false
	}

	// react-native-web is what actually renders the app for the browser; without
	// it there is no web target to export.
	if !p.hasDependency("react-native-web") {
		return false
	}

	// Expo's web output modes are documented at
	// https://docs.expo.dev/versions/latest/config/app/#output.
	output := strings.ToLower(p.getExpoWebOutput(ctx))
	return output == "static" || output == "single"
}

// getExpoWebOutput resolves the configured expo.web.output value from app.json.
// Dynamic app.config.js/ts configs are intentionally not parsed since their
// values can be computed at runtime; those projects can fall back to the
// SPA_OUTPUT_DIR config variable.
func (p *NodeProvider) getExpoWebOutput(ctx *generate.GenerateContext) string {
	if !ctx.App.HasFile("app.json") {
		return ""
	}

	var config expoAppConfig
	if err := ctx.App.ReadJSON("app.json", &config); err != nil {
		return ""
	}

	return config.Expo.Web.Output
}

func (p *NodeProvider) getExpoOutputDirectory(ctx *generate.GenerateContext) string {
	return DefaultExpoOutputDirectory
}
