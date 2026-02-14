package php

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/config"
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

func TestPhpProviderConfigFromFile(t *testing.T) {
	t.Run("php extensions from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/php-laravel-12-react")
		clearConfigVariable(ctx, "PHP_EXTENSIONS")
		setConfigFromJSON(t, ctx, `{
			"php": {
				"extensions": ["xdebug", "imagick"]
			}
		}`)

		provider := PhpProvider{}
		extensions := provider.getPhpExtensions(ctx)

		require.Contains(t, extensions, "xdebug")
		require.Contains(t, extensions, "imagick")
	})

	t.Run("php env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/php-laravel-12-react")
		clearConfigVariable(ctx, "PHP_EXTENSIONS")
		ctx.Env.SetVariable("RAILPACK_PHP_EXTENSIONS", "imagick")
		setConfigFromJSON(t, ctx, `{
			"php": {
				"extensions": ["xdebug"]
			}
		}`)

		provider := PhpProvider{}
		extensions := provider.getPhpExtensions(ctx)

		require.Contains(t, extensions, "imagick")
		require.NotContains(t, extensions, "xdebug")
	})
}

func setConfigFromJSON(t *testing.T, ctx *generate.GenerateContext, configJSON string) {
	t.Helper()

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(configJSON), &cfg))
	ctx.Config = &cfg
}

func clearConfigVariable(ctx *generate.GenerateContext, variableName string) {
	delete(ctx.Env.Variables, ctx.Env.ConfigVariable(variableName))
}
