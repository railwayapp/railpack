package plan

import (
	"strings"

	"github.com/moby/patternmatcher/ignorefile"
	"github.com/railwayapp/railpack/core/app"
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
	parsed   bool
	patterns []string
	app      *app.App
}

func NewDockerignoreContext(app *app.App) *DockerignoreContext {
	return &DockerignoreContext{
		app: app,
	}
}

// Parse parses the .dockerignore file and caches the results
func (d *DockerignoreContext) Parse() ([]string, error) {
	if !d.parsed {
		patterns, err := CheckAndParseDockerignore(d.app)
		if err != nil {
			return nil, err
		}

		d.patterns = patterns
		d.parsed = true
	}

	return d.patterns, nil
}

func (d *DockerignoreContext) ParseWithLogging(logger interface{ LogInfo(string, ...interface{}) }) ([]string, error) {
	patterns, err := d.Parse()
	if err != nil {
		return nil, err
	}

	if patterns != nil {
		logger.LogInfo("Found .dockerignore file, applying filters")
	}

	return patterns, nil
}
