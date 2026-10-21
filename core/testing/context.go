// testing utilities
package testing

import (
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/logger"
)

// CreateGenerateContext creates a new GenerateContext for testing purposes
func CreateGenerateContext(t *testing.T, path string) *generate.GenerateContext {
	t.Helper() // This marks the function as a test helper, which improves test output

	userApp, err := app.NewApp(path)
	if err != nil {
		t.Fatalf("error creating app: %v", err)
	}

	env := app.NewEnvironment(nil)

	config := config.EmptyConfig()

	ctx, err := generate.NewGenerateContext(userApp, env, config, logger.NewLogger())
	if err != nil {
		t.Fatalf("error creating generate context: %v", err)
	}

	return ctx
}
