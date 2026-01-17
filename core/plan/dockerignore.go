package plan

import (
	"strings"

	"github.com/moby/patternmatcher/ignorefile"
	"github.com/railwayapp/railpack/core/app"
)

// checks if a .dockerignore file exists in the app directory and parses it
func CheckAndParseDockerignore(app *app.App) ([]string, []string, error) {
	if !app.HasFile(".dockerignore") {
		return nil, nil, nil
	}

	content, err := app.ReadFile(".dockerignore")
	if err != nil {
		return nil, nil, err
	}

	reader := strings.NewReader(content)
	patterns, err := ignorefile.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	excludePatterns, includePatterns := separatePatterns(patterns)

	// Filter include patterns that don't match any files
	// This prevents build errors when a negated pattern in .dockerignore refers to a missing file
	var validIncludes []string
	for _, pattern := range includePatterns {
		if app.HasMatch(pattern) {
			validIncludes = append(validIncludes, pattern)
		}
	}

	return excludePatterns, validIncludes, nil
}

// separatePatterns separates patterns into exclude and include lists
// Include patterns are those starting with '!' (negation)
func separatePatterns(patterns []string) (excludes []string, includes []string) {
	for _, pattern := range patterns {
		if len(pattern) > 0 && pattern[0] == '!' {
			// Remove the '!' prefix for include patterns
			includes = append(includes, pattern[1:])
		} else {
			excludes = append(excludes, pattern)
		}
	}
	return excludes, includes
}

type DockerignoreContext struct {
	parsed   bool
	excludes []string
	includes []string
	app      *app.App
}

func NewDockerignoreContext(app *app.App) *DockerignoreContext {
	return &DockerignoreContext{
		app: app,
	}
}

// Parse parses the .dockerignore file and caches the results
func (d *DockerignoreContext) Parse() ([]string, []string, error) {
	if !d.parsed {
		excludes, includes, err := CheckAndParseDockerignore(d.app)
		if err != nil {
			return nil, nil, err
		}

		d.excludes = excludes
		d.includes = includes
		d.parsed = true
	}

	return d.excludes, d.includes, nil
}

func (d *DockerignoreContext) ParseWithLogging(logger interface{ LogInfo(string, ...interface{}) }) ([]string, []string, error) {
	excludes, includes, err := d.Parse()
	if err != nil {
		return nil, nil, err
	}

	if excludes != nil || includes != nil {
		logger.LogInfo("Found .dockerignore file, applying filters")
	}

	return excludes, includes, nil
}
