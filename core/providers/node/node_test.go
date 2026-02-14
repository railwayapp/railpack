package node

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
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

func TestNodeProviderConfigFromFile(t *testing.T) {
	t.Run("node version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createNodeApp(t, map[string]string{
			"package.json": `{"name":"app","scripts":{"start":"node index.js"}}`,
			"index.js":     `console.log("ok")`,
		}))
		testingUtils.ClearConfigVariable(ctx, "NODE_VERSION")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"version":"20.11.0"}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		nodeVersion := ctx.Resolver.Get("node")
		require.Equal(t, "20.11.0", nodeVersion.Version)
		require.Equal(t, "node.version", nodeVersion.Source)
	})

	t.Run("node env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createNodeApp(t, map[string]string{
			"package.json": `{"name":"app","scripts":{"start":"node index.js"}}`,
			"index.js":     `console.log("ok")`,
		}))
		testingUtils.ClearConfigVariable(ctx, "NODE_VERSION")
		ctx.Env.SetVariable("RAILPACK_NODE_VERSION", "21.2.0")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"version":"20.11.0"}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		nodeVersion := ctx.Resolver.Get("node")
		require.Equal(t, "21.2.0", nodeVersion.Version)
		require.Equal(t, "RAILPACK_NODE_VERSION", nodeVersion.Source)
	})

	t.Run("bun version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createNodeApp(t, map[string]string{
			"package.json": `{"name":"app","scripts":{"start":"bun run index.js"}}`,
			"index.js":     `console.log("ok")`,
		}))
		testingUtils.ClearConfigVariable(ctx, "BUN_VERSION")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"bunVersion":"1.1.5"}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		bunVersion := ctx.Resolver.Get("bun")
		require.Equal(t, "1.1.5", bunVersion.Version)
		require.Equal(t, "node.bunVersion", bunVersion.Source)
	})

	t.Run("bun env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createNodeApp(t, map[string]string{
			"package.json": `{"name":"app","scripts":{"start":"bun run index.js"}}`,
			"index.js":     `console.log("ok")`,
		}))
		testingUtils.ClearConfigVariable(ctx, "BUN_VERSION")
		ctx.Env.SetVariable("RAILPACK_BUN_VERSION", "1.2.3")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"bunVersion":"1.1.5"}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		bunVersion := ctx.Resolver.Get("bun")
		require.Equal(t, "1.2.3", bunVersion.Version)
		require.Equal(t, "RAILPACK_BUN_VERSION", bunVersion.Source)
	})

	t.Run("spa output dir from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-vite-react")
		testingUtils.ClearConfigVariable(ctx, OUTPUT_DIR_VAR)
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"spaOutputDir":"custom-dist"}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.True(t, provider.isSPA(ctx))
		require.Equal(t, "custom-dist", provider.getOutputDirectory(ctx))
	})

	t.Run("spa output dir env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-vite-react")
		testingUtils.ClearConfigVariable(ctx, OUTPUT_DIR_VAR)
		ctx.Env.SetVariable("RAILPACK_SPA_OUTPUT_DIR", "env-dist")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"spaOutputDir":"custom-dist"}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.Equal(t, "env-dist", provider.getOutputDirectory(ctx))
	})

	t.Run("no spa from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-vite-react")
		testingUtils.ClearConfigVariable(ctx, "NO_SPA")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"noSpa":true}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.False(t, provider.isSPA(ctx))
	})

	t.Run("prune deps from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-npm")
		testingUtils.ClearConfigVariable(ctx, "PRUNE_DEPS")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"pruneDeps":true}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.True(t, provider.shouldPrune(ctx))
	})

	t.Run("prune command from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-npm")
		testingUtils.ClearConfigVariable(ctx, "NODE_PRUNE_CMD")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"pruneCmd":"npm prune --omit=dev --ignore-scripts --workspaces"}}`)

		prune := ctx.NewCommandStep("prune")
		PackageManagerNpm.PruneDeps(ctx, prune)

		require.Contains(t, getExecCommands(prune), "npm prune --omit=dev --ignore-scripts --workspaces")
	})

	t.Run("prune command env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-npm")
		testingUtils.ClearConfigVariable(ctx, "NODE_PRUNE_CMD")
		ctx.Env.SetVariable("RAILPACK_NODE_PRUNE_CMD", "npm prune --omit=dev")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"pruneCmd":"npm prune --omit=dev --ignore-scripts --workspaces"}}`)

		prune := ctx.NewCommandStep("prune")
		PackageManagerNpm.PruneDeps(ctx, prune)

		require.Contains(t, getExecCommands(prune), "npm prune --omit=dev")
	})

	t.Run("install patterns from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createNodeApp(t, map[string]string{
			"package.json":             `{"name":"app"}`,
			"custom/from-config.txt":   "configured",
			"custom/from-env-only.txt": "env",
		}))
		testingUtils.ClearConfigVariable(ctx, "NODE_INSTALL_PATTERNS")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"installPatterns":["from-config.txt"]}}`)

		files := PackageManagerNpm.SupportingInstallFiles(ctx)
		require.Contains(t, files, "custom/from-config.txt")
	})

	t.Run("install patterns env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createNodeApp(t, map[string]string{
			"package.json":             `{"name":"app"}`,
			"custom/from-config.txt":   "configured",
			"custom/from-env-only.txt": "env",
		}))
		testingUtils.ClearConfigVariable(ctx, "NODE_INSTALL_PATTERNS")
		ctx.Env.SetVariable("RAILPACK_NODE_INSTALL_PATTERNS", "from-env-only.txt")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"installPatterns":["from-config.txt"]}}`)

		files := PackageManagerNpm.SupportingInstallFiles(ctx)
		require.Contains(t, files, "custom/from-env-only.txt")
		require.NotContains(t, files, "custom/from-config.txt")
	})

	t.Run("angular project from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createAngularWorkspaceApp(t))
		testingUtils.ClearConfigVariable(ctx, "ANGULAR_PROJECT")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"angularProject":"admin"}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.Equal(t, "dist/admin/browser", provider.getAngularOutputDirectory(ctx))
	})

	t.Run("angular project env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createAngularWorkspaceApp(t))
		testingUtils.ClearConfigVariable(ctx, "ANGULAR_PROJECT")
		ctx.Env.SetVariable("RAILPACK_ANGULAR_PROJECT", "web")
		testingUtils.SetConfigFromJSON(t, ctx, `{"node":{"angularProject":"admin"}}`)

		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.Equal(t, "dist/web/browser", provider.getAngularOutputDirectory(ctx))
	})
}

func getExecCommands(step *generate.CommandStepBuilder) []string {
	commands := []string{}

	for _, command := range step.Commands {
		execCommand, ok := command.(plan.ExecCommand)
		if ok {
			commands = append(commands, execCommand.Cmd)
		}
	}

	return commands
}

func createNodeApp(t *testing.T, files map[string]string) string {
	t.Helper()

	appDir := t.TempDir()

	for filePath, contents := range files {
		absolutePath := filepath.Join(appDir, filePath)
		require.NoError(t, os.MkdirAll(filepath.Dir(absolutePath), 0755))
		require.NoError(t, os.WriteFile(absolutePath, []byte(contents), 0644))
	}

	return appDir
}

func createAngularWorkspaceApp(t *testing.T) string {
	t.Helper()

	return createNodeApp(t, map[string]string{
		"package.json": `{
			"name": "angular-app",
			"dependencies": {"@angular/core": "^19.0.0"},
			"scripts": {"build": "ng build"}
		}`,
		"angular.json": `{
			"projects": {
				"web": {
					"architect": {
						"build": {
							"builder": "@angular-devkit/build-angular:application",
							"options": {"outputPath": "dist/web"}
						}
					}
				},
				"admin": {
					"architect": {
						"build": {
							"builder": "@angular-devkit/build-angular:application",
							"options": {"outputPath": "dist/admin"}
						}
					}
				}
			}
		}`,
	})
}
