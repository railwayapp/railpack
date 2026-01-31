package plan

import (
	"strings"

	"github.com/railwayapp/railpack/core/app"

	// this is the native dockerignore parser used by buildkit, it uses patternmatcher in its implementation
	// https://github.com/moby/buildkit/blob/master/frontend/dockerfile/dockerignore/dockerignore_deprecated.go
	"github.com/moby/patternmatcher/ignorefile"
)

// checks if a .dockerignore file exists in the app directory and parses it
func CheckAndParseDockerignore(app *app.App) ([]string, error) {
	if !app.HasFile(".dockerignore") {
		return nil, nil
	}

	content, err := app.ReadFile(".dockerignore")
	if err != nil {
		return nil, err
	}

	reader := strings.NewReader(content)
	patterns, err := ignorefile.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return patterns, nil
}

type DockerignoreContext struct {
	Excludes []string
	HasFile  bool
}

func NewDockerignoreContext(app *app.App) (*DockerignoreContext, error) {
	hasFile := app.HasFile(".dockerignore")
	excludes, err := CheckAndParseDockerignore(app)
	if err != nil {
		return nil, err
	}
	return &DockerignoreContext{
		Excludes: excludes,
		HasFile:  hasFile,
	}, nil
}
