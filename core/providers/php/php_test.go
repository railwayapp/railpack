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

func TestPhpAptPackagesSpread(t *testing.T) {
	t.Run("spread entry expands instead of leaking into apt-get", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/php-vanilla")
		ctx.Config.BuildAptPackages = []string{"...", "curl"}
		ctx.Config.Deploy.AptPackages = []string{"...", "telnet"}
		provider := PhpProvider{}

		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		var aptPackages []string
		for _, step := range ctx.Steps {
			if imageStep, ok := step.(*generate.ImageStepBuilder); ok {
				aptPackages = append(aptPackages, imageStep.AptPackages...)
			}
		}
		require.NotEmpty(t, aptPackages)
		require.NotContains(t, aptPackages, "...")
		require.Contains(t, aptPackages, "curl")
		require.Contains(t, aptPackages, "telnet")
		require.Contains(t, aptPackages, "git")
	})

	t.Run("build packages without spread keep additive behavior", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/php-vanilla")
		ctx.Config.BuildAptPackages = []string{"curl"}
		provider := PhpProvider{}

		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		var aptPackages []string
		for _, step := range ctx.Steps {
			if imageStep, ok := step.(*generate.ImageStepBuilder); ok {
				aptPackages = append(aptPackages, imageStep.AptPackages...)
			}
		}
		require.Contains(t, aptPackages, "curl")
		require.Contains(t, aptPackages, "git")
	})
}
