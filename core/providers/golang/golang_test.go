package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/plan"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGolang(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		detected     bool
		hasGoMod     bool
		hasWorkspace bool
		goVersion    string
		cgoEnabled   bool
	}{
		{
			name:      "go mod",
			path:      "../../../examples/go-mod",
			detected:  true,
			hasGoMod:  true,
			goVersion: "1.25.3",
		},
		{
			name:      "go cmd dirs",
			path:      "../../../examples/go-cmd-dirs",
			detected:  true,
			hasGoMod:  true,
			goVersion: "1.25.3",
		},
		{
			name:         "go workspaces",
			path:         "../../../examples/go-workspaces",
			detected:     true,
			hasGoMod:     false,
			hasWorkspace: true,
			goVersion:    "1.25",
		},
		{
			name:     "node",
			path:     "../../../examples/node-npm",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := GoProvider{}
			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				err = provider.Plan(ctx)
				require.NoError(t, err)

				require.Equal(t, tt.hasGoMod, provider.isGoMod(ctx))
				require.Equal(t, tt.hasWorkspace, provider.isGoWorkspace(ctx))
				require.Equal(t, tt.cgoEnabled, provider.hasCGOEnabled(ctx))

				if tt.goVersion != "" {
					goVersion := ctx.Resolver.Get("go")
					require.Equal(t, tt.goVersion, goVersion.Version)
				}

				if tt.hasWorkspace {
					packages := provider.GoWorkspacePackages(ctx)
					require.Greater(t, len(packages), 0, "workspace should have at least one package")
				}
			}
		})
	}
}

func buildCommands(t *testing.T, files map[string]string) []string {
	t.Helper()

	dir := t.TempDir()
	for name, contents := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	}

	ctx := testingUtils.CreateGenerateContext(t, dir)
	provider := GoProvider{}
	require.NoError(t, provider.Plan(ctx))

	buildPlan, _, err := ctx.Generate()
	require.NoError(t, err)

	cmds := []string{}
	for _, step := range buildPlan.Steps {
		if step.Name != "build" {
			continue
		}
		for _, cmd := range step.Commands {
			if exec, ok := cmd.(plan.ExecCommand); ok {
				cmds = append(cmds, exec.Cmd)
			}
		}
	}
	return cmds
}

func TestGolangBuildsCmdDirWhenRootPackageIsALibrary(t *testing.T) {
	cmds := buildCommands(t, map[string]string{
		"go.mod":             "module example.com/app\n\ngo 1.25\n",
		"lib.go":             "package app\n\nfunc Add(a, b int) int { return a + b }\n",
		"lib_test.go":        "package main\n",
		"cmd/server/main.go": "package main\n\nfunc main() {}\n",
	})
	require.Equal(t, []string{`go build -ldflags="-w -s" -o out ./cmd/server`}, cmds)
}

func TestGolangBuildsRootWhenRootPackageIsMain(t *testing.T) {
	cmds := buildCommands(t, map[string]string{
		"go.mod":             "module example.com/app\n\ngo 1.25\n",
		"main.go":            "// Command app does things.\npackage main\n\nfunc main() {}\n",
		"cmd/server/main.go": "package main\n\nfunc main() {}\n",
	})
	require.Equal(t, []string{`go build -ldflags="-w -s" -o out`}, cmds)
}
