package node

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DefaultNextOutputDirectory = "out"
	DefaultNextStartCommand    = "next start"
)

var nextConfigFiles = []string{
	"next.config.js",
	"next.config.mjs",
	"next.config.ts",
}

func (p *NodeProvider) isNextSPA(ctx *generate.GenerateContext) bool {
	if !p.isNext() {
		return false
	}

	configFileContents := p.getNextConfigFileContents(ctx)
	hasExportOutput := strings.Contains(configFileContents, "output: 'export'") || strings.Contains(configFileContents, "output: \"export\"")

	return hasExportOutput
}

func (p *NodeProvider) getNextOutputDirectory(ctx *generate.GenerateContext) string {
	configFileContents := p.getNextConfigFileContents(ctx)
	if configFileContents != "" {
		distDirRegex := regexp.MustCompile(`distDir:\s*['"](.+?)['"]`)
		matches := distDirRegex.FindStringSubmatch(configFileContents)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return DefaultNextOutputDirectory
}

func (p *NodeProvider) getNextConfigFileContents(ctx *generate.GenerateContext) string {
	_, contents, _ := ctx.App.ReadFirstFileOf(nextConfigFiles...)
	return contents
}
