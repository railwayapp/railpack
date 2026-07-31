package plan

import (
	"strings"

	"github.com/railwayapp/railpack/core/app"

	// this is the native dockerignore parser used by buildkit
	// https://github.com/moby/buildkit/blob/master/frontend/dockerfile/dockerignore/dockerignore_deprecated.go
	"github.com/moby/patternmatcher/ignorefile"
)

// parses a .dockerignore file from the app directory. assumes the file exists.
func checkAndParseDockerignore(app *app.App) ([]string, error) {
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
	if !app.HasFile(".dockerignore") {
		return &DockerignoreContext{}, nil
	}

	excludes, err := checkAndParseDockerignore(app)
	if err != nil {
		return nil, err
	}

	return &DockerignoreContext{
		Excludes: excludes,
		HasFile:  true,
	}, nil
}
