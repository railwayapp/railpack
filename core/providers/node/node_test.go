package node

import (
	"fmt"
	"strings"
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestNode(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		detected       bool
		packageManager PackageManager
		nodeVersion    string
		pnpmVersion    string
	}{
		{
			name:           "npm",
			path:           "../../../examples/node-npm",
			detected:       true,
			packageManager: PackageManagerNpm,
			nodeVersion:    "23.5.0",
		},
		{
			name:           "bun",
			path:           "../../../examples/node-bun",
			detected:       true,
			packageManager: PackageManagerBun,
		},
		{
			name:           "pnpm",
			path:           "../../../examples/node-corepack",
			detected:       true,
			packageManager: PackageManagerPnpm,
			nodeVersion:    "20",
			pnpmVersion:    "10.4.1",
		},
		{
			name:           "pnpm",
			path:           "../../../examples/node-pnpm-workspaces",
			detected:       true,
			packageManager: PackageManagerPnpm,
			nodeVersion:    "22.2.0",
		},
		{
			name:           "pnpm",
			path:           "../../../examples/node-astro",
			detected:       true,
			packageManager: PackageManagerNpm,
		},
		{
			name:     "golang",
			path:     "../../../examples/go-mod",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}
			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				packageManager := provider.getPackageManager(ctx.App)
				require.Equal(t, tt.packageManager, packageManager)

				err = provider.Plan(ctx)
				require.NoError(t, err)

				if tt.nodeVersion != "" {
					nodeVersion := ctx.Resolver.Get("node")
					require.Equal(t, tt.nodeVersion, nodeVersion.Version)
				}

				if tt.pnpmVersion != "" {
					pnpmVersion := ctx.Resolver.Get("pnpm")
					require.Equal(t, tt.pnpmVersion, pnpmVersion.Version)
				}
			}
		})
	}
}

func TestNodeCorepack(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantCorepack bool
	}{
		{
			name:         "corepack project",
			path:         "../../../examples/node-corepack",
			wantCorepack: true,
		},
		{
			name:         "bun project",
			path:         "../../../examples/node-bun",
			wantCorepack: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}
			err := provider.Initialize(ctx)
			require.NoError(t, err)

			usesCorepack := provider.usesCorepack()
			require.Equal(t, tt.wantCorepack, usesCorepack)
		})
	}
}

func TestGetNextApps(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "npm project",
			path: "../../../examples/node-npm",
			want: []string{},
		},
		{
			name: "bun project",
			path: "../../../examples/node-next",
			want: []string{""},
		},
		{
			name: "turbo with 2 next apps",
			path: "../../../examples/node-turborepo",
			want: []string{"apps/web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}
			err := provider.Initialize(ctx)
			require.NoError(t, err)

			nextPackages, err := provider.getPackagesWithFramework(ctx, func(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
				if pkg.PackageJson.HasScript("build") {
					return strings.Contains(pkg.PackageJson.Scripts["build"], "next build")
				}
				return false
			})
			require.NoError(t, err)

			nextApps := make([]string, len(nextPackages))
			for i, pkg := range nextPackages {
				nextApps[i] = pkg.Path
			}
			require.Equal(t, tt.want, nextApps)
		})
	}
}

func TestPackageJsonRequiresBun(t *testing.T) {
	// Special cases
	t.Run("nil package.json", func(t *testing.T) {
		got := packageJsonRequiresBun(nil)
		require.False(t, got)
	})

	t.Run("no scripts", func(t *testing.T) {
		got := packageJsonRequiresBun(&PackageJson{})
		require.False(t, got)
	})

	// Scripts that should trigger bun detection
	bunScripts := []string{
		"bun run server.js",
		"bunx nodemon index.js",
		"bun test",
		"npm run clean && bun build.js",
		"echo 'Running tests' | bun test",
		"npm run build; bun run server.js",
		"cd src && bun install",
		"bun --version",
		"bunx prisma migrate",
	}

	t.Run("scripts requiring bun", func(t *testing.T) {
		packageJson := &PackageJson{
			Scripts: make(map[string]string),
		}
		for i, script := range bunScripts {
			packageJson.Scripts[fmt.Sprintf("script%d", i)] = script
		}
		got := packageJsonRequiresBun(packageJson)
		require.True(t, got)
	})

	// Scripts that should NOT trigger bun detection
	nonBunScripts := []string{
		"esbuild dev.ts ./src --bundle --outdir=dist --packages=external --platform=node --sourcemap --watch",
		"webpack --config webpack.bundle.config.js",
		"node src/bundle-manager.js",
		"jest --bundle-reporter",
		"eslint src/bundles/",
		"sh deploy-bundle.sh",
		"npm run bundle:production",
		"yarn bundle",
		"pnpm run unbundle",
	}

	t.Run("scripts not requiring bun", func(t *testing.T) {
		packageJson := &PackageJson{
			Scripts: make(map[string]string),
		}
		for i, script := range nonBunScripts {
			packageJson.Scripts[fmt.Sprintf("script%d", i)] = script
		}
		got := packageJsonRequiresBun(packageJson)
		require.False(t, got)
	})
}
