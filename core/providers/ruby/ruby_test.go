package ruby

import (
	"testing"

	"github.com/Masterminds/semver/v3"
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

func TestRubyVersionDetection_Examples(t *testing.T) {
	cases := []struct {
		desc   string
		path   string
		expect string
	}{
		{
			desc:   ".ruby-version present (ruby-3)",
			path:   "../../../examples/ruby-3",
			expect: "3.2.1",
		},
		{
			desc:   "Gemfile.lock with version (ruby-vanilla)",
			path:   "../../../examples/ruby-vanilla",
			expect: "3.4.2",
		},
		{
			desc:   "Gemfile with version constraint < 3.5.0 (ruby-no-version)",
			path:   "../../../examples/ruby-no-version",
			expect: "3.4.4",
		},
		{
			desc:   "Gemfile with single version constraint (ruby-single-version)",
			path:   "../../../examples/ruby-single-version",
			expect: "3.4.2",
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, c.path)
			miseStep := ctx.GetMiseStepBuilder()
			provider := RubyProvider{}
			provider.InstallMisePackages(ctx, miseStep)

			resolvedPackages, err := miseStep.Resolver.ResolvePackages()
			require.NoError(t, err)

			resolvedPkg, exists := resolvedPackages["ruby"]
			if !exists {
				t.Fatalf("ruby package not found in resolved packages")
			}

			if resolvedPkg.ResolvedVersion == nil {
				t.Fatalf("ruby package version not resolved")
			}

			require.Equal(t, c.expect, *resolvedPkg.ResolvedVersion)
		})
	}
}

func TestRubyVersionConstraintResolution(t *testing.T) {
	t.Run("constraint_resolution_works", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/ruby-no-version")

		miseStep := ctx.GetMiseStepBuilder()
		provider := RubyProvider{}
		provider.InstallMisePackages(ctx, miseStep)

		resolvedPackages, err := miseStep.Resolver.ResolvePackages()
		require.NoError(t, err)

		resolvedPkg, exists := resolvedPackages["ruby"]
		require.True(t, exists, "ruby package should be found in resolved packages")
		require.NotNil(t, resolvedPkg.ResolvedVersion, "ruby package version should be resolved")

		actualVersion := *resolvedPkg.ResolvedVersion

		v, err := semver.NewVersion(actualVersion)
		require.NoError(t, err)
		require.True(t, v.Major() == 3 && v.Minor() >= 2 && v.Minor() < 5,
			"Resolved version %s should be >= 3.2.0 and < 3.5.0", actualVersion)

		require.True(t, v.Minor() >= 4, "Should resolve to a recent version, got %s", actualVersion)

		require.Equal(t, "3.4.4", actualVersion, "Should resolve to the highest stable version under 3.5.0")
	})
}
