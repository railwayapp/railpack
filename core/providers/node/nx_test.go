package node

import (
	"testing"

	"github.com/railwayapp/railpack/core/plan"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestIsNx(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "nx next workspace",
			path: "../../../examples/node-nx-next",
			want: true,
		},
		{
			name: "plain next app",
			path: "../../../examples/node-next",
			want: false,
		},
		{
			name: "turborepo",
			path: "../../../examples/node-turborepo",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}
			err := provider.Initialize(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, provider.isNx(ctx))
		})
	}
}

func TestResolveNxDeployPackage(t *testing.T) {
	t.Run("single next app in nx workspace", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-nx-next")
		provider := NodeProvider{}
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		pkg, name, ok := provider.resolveNxDeployPackage(ctx)
		require.True(t, ok)
		require.Equal(t, "apps/web", pkg.Path)
		require.Equal(t, "@node-nx-next/web", name)
	})

	t.Run("RAILPACK_NX_APP selects by short name", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-nx-next")
		ctx.Env.Variables["RAILPACK_NX_APP"] = "web"
		provider := NodeProvider{}
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		pkg, name, ok := provider.resolveNxDeployPackage(ctx)
		require.True(t, ok)
		require.Equal(t, "apps/web", pkg.Path)
		require.Equal(t, "@node-nx-next/web", name)
	})

	t.Run("RAILPACK_NX_APP selects by package name", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-nx-next")
		ctx.Env.Variables["RAILPACK_NX_APP"] = "@node-nx-next/web"
		provider := NodeProvider{}
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		pkg, name, ok := provider.resolveNxDeployPackage(ctx)
		require.True(t, ok)
		require.Equal(t, "apps/web", pkg.Path)
		require.Equal(t, "@node-nx-next/web", name)
	})

	t.Run("non-nx project", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-next")
		provider := NodeProvider{}
		err := provider.Initialize(ctx)
		require.NoError(t, err)

		_, _, ok := provider.resolveNxDeployPackage(ctx)
		require.False(t, ok)
	})
}

func TestNxBuildAndStartCommands(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-nx-next")
	provider := NodeProvider{}
	err := provider.Initialize(ctx)
	require.NoError(t, err)

	require.Equal(t, "cd apps/web && next start", provider.GetStartCommand(ctx))

	// Build step should use nx when root has no build script
	build := ctx.NewCommandStep("build")
	provider.Build(ctx, build)

	var foundNxBuild bool
	for _, cmd := range build.Commands {
		if execCmd, ok := cmd.(plan.ExecCommand); ok && execCmd.Cmd == "nx build @node-nx-next/web" {
			foundNxBuild = true
			break
		}
	}
	require.True(t, foundNxBuild, "expected nx build command in build step")
	require.Equal(t, "1", build.Variables["NEXT_TELEMETRY_DISABLED"])
}

func TestNxDoesNotOverrideRootScripts(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-turborepo")
	provider := NodeProvider{}
	err := provider.Initialize(ctx)
	require.NoError(t, err)

	// turborepo has root start/build scripts and is not Nx
	require.False(t, provider.isNx(ctx))
	require.Equal(t, "npm run start", provider.GetStartCommand(ctx))
}

func TestMatchesNxAppSelector(t *testing.T) {
	pkg := &WorkspacePackage{
		Path: "apps/web",
		PackageJson: &PackageJson{
			Name: "@node-nx-next/web",
		},
	}

	require.True(t, matchesNxAppSelector(pkg, "@node-nx-next/web"))
	require.True(t, matchesNxAppSelector(pkg, "apps/web"))
	require.True(t, matchesNxAppSelector(pkg, "web"))
	require.False(t, matchesNxAppSelector(pkg, "docs"))
	require.False(t, matchesNxAppSelector(pkg, ""))
}
