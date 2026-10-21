package php

import (
	"strings"
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestPhpProvider(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		isPhp     bool
		framework string
	}{
		{
			name:      "vanilla php with index.php",
			path:      "../../../examples/php-vanilla",
			isPhp:     true,
			framework: "",
		},
		{
			name:      "laravel project with composer.json",
			path:      "../../../examples/php-laravel-12-react",
			isPhp:     true,
			framework: "laravel",
		},
		{
			name:      "symfony project",
			path:      "../../../examples/php-symfony",
			isPhp:     true,
			framework: "symfony",
		},
		{
			name:      "non-php project",
			path:      "../../../examples/node-npm",
			isPhp:     false,
			framework: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PhpProvider{}

			isPhp, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.isPhp, isPhp)

			require.Equal(t, tt.framework, provider.getFramework(ctx))
		})
	}
}

// PHP installs build and deploy apt packages on one image, so "..." is removed and the lists are merged.
func TestPhpCombinesBuildAndDeployAptPackages(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/php-vanilla")
	ctx.Config.BuildAptPackages = []string{"...", "curl", "git"}
	ctx.Config.Deploy.AptPackages = []string{"...", "curl", "jq"}

	provider := PhpProvider{}
	require.NoError(t, provider.Initialize(ctx))
	require.NoError(t, provider.Plan(ctx))

	var packages []string
	for _, step := range ctx.Steps {
		imageStep, ok := step.(*generate.ImageStepBuilder)
		if !ok {
			continue
		}
		packages = imageStep.AptPackages
	}

	require.Equal(t, []string{"git", "zip", "unzip", "ca-certificates", "curl", "jq"}, packages)
}

// Laravel caches must be warmed at container startup (with config:clear before migrate) rather than during the image build.
func TestLaravelDefersCacheWarmingToStartupScript(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/php-laravel-12-react")

	provider := PhpProvider{}
	require.NoError(t, provider.Initialize(ctx))
	require.NoError(t, provider.Plan(ctx))

	var startScript string
	for _, step := range ctx.Steps {
		cmdStep, ok := step.(*generate.CommandStepBuilder)
		if !ok {
			continue
		}
		if cmdStep.Name() == "prepare" {
			startScript = cmdStep.Assets["start-container.sh"]
		}
		if cmdStep.Name() == "build" {
			for _, cmd := range cmdStep.Commands {
				execCmd, ok := cmd.(plan.ExecCommand)
				if !ok {
					continue
				}
				require.NotContains(t, execCmd.Cmd, "php artisan")
			}
		}
	}

	require.NotEmpty(t, startScript)
	require.NotContains(t, startScript, "php artisan optimize:clear")

	clearIdx := strings.Index(startScript, "php artisan config:clear")
	migrateIdx := strings.Index(startScript, "php artisan migrate --force")
	optimizeIdx := strings.Index(startScript, "php artisan optimize")
	require.Greater(t, clearIdx, -1)
	require.Greater(t, migrateIdx, clearIdx)
	require.Greater(t, optimizeIdx, migrateIdx)
}
