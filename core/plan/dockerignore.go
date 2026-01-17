package plan

import (
	"strings"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/internal/utils"

	// this is the native dockerignore parser used by buildkit
	// https://github.com/moby/buildkit/blob/master/frontend/dockerfile/dockerignore/dockerignore_deprecated.go
	"github.com/moby/patternmatcher/ignorefile"
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
	Excludes []string
	Includes []string
	HasFile  bool
}

func NewDockerignoreContext(app *app.App) (*DockerignoreContext, error) {
	hasFile := app.HasFile(".dockerignore")
	excludes, includes, err := CheckAndParseDockerignore(app)
	if err != nil {
		return nil, err
	}
	if excludes != nil {
		excludes = utils.RemoveDuplicates(excludes)
	}
	if includes != nil {
		includes = utils.RemoveDuplicates(includes)
	}
	return &DockerignoreContext{
		Excludes: excludes,
		Includes: includes,
		HasFile:  hasFile,
	}, nil
}
