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

func TestGolangBuildCommand(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		envs        map[string]string
		expectedCmd string
	}{
		{
			name: "builds cmd dir when root package is a library",
			files: map[string]string{
				"go.mod":             "module example.com/app\n\ngo 1.25\n",
				"lib.go":             "package app\n\nfunc Add(a, b int) int { return a + b }\n",
				"lib_test.go":        "package app_test\n",
				"cmd/server/main.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out ./cmd/server`,
		},
		{
			name: "ignores package main inside comments and strings in root library",
			files: map[string]string{
				"go.mod":             "module example.com/app\n\ngo 1.25\n",
				"lib.go":             "/*\npackage main\n*/\npackage app\n\nconst example = `\npackage main\n`\n",
				"cmd/server/main.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out ./cmd/server`,
		},
		{
			name: "ignores build-constrained ignore helper in root",
			files: map[string]string{
				"go.mod":             "module example.com/app\n\ngo 1.25\n",
				"lib.go":             "package app\n",
				"generate.go":        "//go:build ignore\n\npackage main\n\nfunc main() {}\n",
				"cmd/server/main.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out ./cmd/server`,
		},
		{
			name: "ignores legacy plus-build ignore helper in root",
			files: map[string]string{
				"go.mod":             "module example.com/app\n\ngo 1.25\n",
				"lib.go":             "package app\n",
				"generate.go":        "// +build ignore\n\npackage main\n\nfunc main() {}\n",
				"cmd/server/main.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out ./cmd/server`,
		},
		{
			name: "ignores non-linux platform-constrained main file in root",
			files: map[string]string{
				"go.mod":             "module example.com/app\n\ngo 1.25\n",
				"lib.go":             "package app\n",
				"main_windows.go":    "//go:build windows\n\npackage main\n\nfunc main() {}\n",
				"cmd/server/main.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out ./cmd/server`,
		},
		{
			name: "builds root when platform constraint matches linux target",
			files: map[string]string{
				"go.mod":             "module example.com/app\n\ngo 1.25\n",
				"main_unix.go":       "//go:build !windows\n\npackage main\n\nfunc main() {}\n",
				"cmd/server/main.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out`,
		},
		{
			name: "skips non-main helper packages under cmd",
			files: map[string]string{
				"go.mod":                "module example.com/app\n\ngo 1.25\n",
				"cmd/common/helpers.go": "package common\n",
				"cmd/server/main.go":    "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out ./cmd/server`,
		},
		{
			name: "builds root when root package is main",
			files: map[string]string{
				"go.mod":             "module example.com/app\n\ngo 1.25\n",
				"main.go":            "// Command app does things.\npackage main\n\nfunc main() {}\n",
				"cmd/server/main.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out`,
		},
		{
			name: "builds workspace main module when root go.mod and earlier workspace modules are libraries",
			files: map[string]string{
				"go.work":          "go 1.25\n\nuse (\n\t.\n\t./common\n\t./server\n)\n",
				"go.mod":           "module example.com/root\n\ngo 1.25\n",
				"root.go":          "package root\n",
				"common/go.mod":    "module example.com/common\n\ngo 1.25\n",
				"common/common.go": "package common\n",
				"server/go.mod":    "module example.com/server\n\ngo 1.25\n",
				"server/server.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out ./server`,
		},
		{
			name: "falls back to building main.go when no go.mod or go.work exists",
			files: map[string]string{
				"main.go": "package main\n\nfunc main() {}\n",
			},
			expectedCmd: `go build -ldflags="-w -s" -o out main.go`,
		},
		{
			name: "RAILPACK_GO_BIN selects specified cmd binary over root main and earlier cmd",
			files: map[string]string{
				"go.mod":             "module example.com/app\n\ngo 1.25\n",
				"main.go":            "package main\n\nfunc main() {}\n",
				"cmd/server/main.go": "package main\n\nfunc main() {}\n",
				"cmd/worker/main.go": "package main\n\nfunc main() {}\n",
			},
			envs:        map[string]string{"RAILPACK_GO_BIN": "worker"},
			expectedCmd: `go build -ldflags="-w -s" -o out ./cmd/worker`,
		},
		{
			name: "RAILPACK_GO_WORKSPACE_MODULE selects specified workspace module over earlier module",
			files: map[string]string{
				"go.work":        "go 1.25\n\nuse (\n\t./api\n\t./worker\n)\n",
				"api/go.mod":     "module example.com/api\n\ngo 1.25\n",
				"api/main.go":    "package main\n\nfunc main() {}\n",
				"worker/go.mod":  "module example.com/worker\n\ngo 1.25\n",
				"worker/main.go": "package main\n\nfunc main() {}\n",
			},
			envs:        map[string]string{"RAILPACK_GO_WORKSPACE_MODULE": "worker"},
			expectedCmd: `go build -ldflags="-w -s" -o out ./worker`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, contents := range tt.files {
				path := filepath.Join(dir, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
			}

			ctx := testingUtils.CreateGenerateContext(t, dir)
			for k, v := range tt.envs {
				ctx.Env.SetVariable(k, v)
			}
			build := ctx.NewCommandStep("build")

			provider := GoProvider{}
			provider.Build(ctx, build)

			require.Equal(t, []plan.Command{plan.NewExecCommand(tt.expectedCmd)}, build.Commands)
		})
	}
}
