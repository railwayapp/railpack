package php

import (
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestPhpProvider(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		isPhp     bool
		isLaravel bool
	}{
		{
			name:      "vanilla php with index.php",
			path:      "../../../examples/php-vanilla",
			isPhp:     true,
			isLaravel: false,
		},
		{
			name:      "laravel project with composer.json",
			path:      "../../../examples/php-laravel-12-react",
			isPhp:     true,
			isLaravel: true,
		},
		{
			name:      "non-php project",
			path:      "../../../examples/node-npm",
			isPhp:     false,
			isLaravel: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PhpProvider{}

			isPhp, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.isPhp, isPhp)

			isLaravel := provider.usesLaravel(ctx)
			require.Equal(t, tt.isLaravel, isLaravel)
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
