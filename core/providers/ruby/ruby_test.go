package ruby

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"

	"github.com/stretchr/testify/require"

	testingUtils "github.com/railwayapp/railpack/core/testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "ruby",
			path: "../../../examples/ruby-vanilla",
			want: true,
		},
		{
			name: "no ruby",
			path: "../../../examples/go-mod",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := RubyProvider{}
			got, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRubyProviderConfigFromFile(t *testing.T) {
	t.Run("ruby version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createRubyApp(t))
		clearConfigVariable(ctx, "RUBY_VERSION")
		setConfigFromJSON(t, ctx, `{
			"ruby": {
				"version": "3.2.4"
			}
		}`)

		provider := RubyProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		rubyVersion := ctx.Resolver.Get("ruby")
		require.Equal(t, "3.2.4", rubyVersion.Version)
		require.Equal(t, "ruby.version", rubyVersion.Source)
	})

	t.Run("ruby env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createRubyApp(t))
		clearConfigVariable(ctx, "RUBY_VERSION")
		ctx.Env.SetVariable("RAILPACK_RUBY_VERSION", "3.3.1")
		setConfigFromJSON(t, ctx, `{
			"ruby": {
				"version": "3.2.4"
			}
		}`)

		provider := RubyProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		rubyVersion := ctx.Resolver.Get("ruby")
		require.Equal(t, "3.3.1", rubyVersion.Version)
		require.Equal(t, "RAILPACK_RUBY_VERSION", rubyVersion.Source)
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

func createRubyApp(t *testing.T) string {
	t.Helper()

	appDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(appDir, "Gemfile"), []byte(`source "https://rubygems.org"`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "Gemfile.lock"), []byte(`GEM
  remote: https://rubygems.org/
  specs:

PLATFORMS
  x86_64-linux

DEPENDENCIES

RUBY VERSION
   ruby 3.1.2p20

BUNDLED WITH
   2.3.7
`), 0644))

	return appDir
}
