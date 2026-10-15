package ruby

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/resolver"
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

func TestRubyVersion(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		version string
		source  string
		env     map[string]string
	}{
		{
			name:    "gemfile lock fallback",
			path:    "../../../examples/ruby-execjs",
			version: "3.1.2",
			source:  "Gemfile.lock",
		},
		{
			name:    "env overrides lockfile",
			path:    "../../../examples/ruby-execjs",
			version: "3.3",
			source:  "RAILPACK_RUBY_VERSION",
			env:     map[string]string{"RAILPACK_RUBY_VERSION": "3.3"},
		},
		{
			name:    "mise config overrides lockfile",
			path:    "../../../examples/ruby-vanilla",
			version: "3.3.0",
			source:  "mise.toml",
		},
		{
			name:    "gemfile ruby directive",
			path:    "../../../examples/ruby-sinatra",
			version: "3.1.2",
			source:  "idiomatic-version-file",
		},
		{
			name:    "gemfile ruby file directive",
			path:    "../../../examples/ruby-rails-with-frontend",
			version: "3.4.1",
			source:  "idiomatic-version-file",
		},
		{
			name:    "default",
			path:    "../../../examples/ruby-no-version",
			version: "3.4",
			source:  resolver.DefaultSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			if tt.env != nil {
				envVars := tt.env
				ctx.Env = app.NewEnvironment(&envVars)
			}

			provider := RubyProvider{}
			provider.InstallMisePackages(ctx, ctx.GetMiseStepBuilder())

			pkg := ctx.Resolver.Get("ruby")
			require.NotNil(t, pkg)
			require.Equal(t, tt.version, pkg.Version)
			require.Equal(t, tt.source, pkg.Source)
		})
	}
}
