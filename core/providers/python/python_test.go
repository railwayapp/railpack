package python

import (
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
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
			name: "pip",
			path: "../../../examples/python-pip",
			want: true,
		},
		{
			name: "poetry",
			path: "../../../examples/python-poetry",
			want: true,
		},
		{
			name: "pdm",
			path: "../../../examples/python-pdm",
			want: true,
		},
		{
			name: "uv",
			path: "../../../examples/python-uv",
			want: true,
		},
		{
			name: "bot.py only",
			path: "../../../examples/python-bot-only",
			want: true,
		},
		{
			name: "no python",
			path: "../../../examples/go-mod",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PythonProvider{}
			got, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUsesBinaryPsycopg(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "psycopg2-binary",
			path: "../../../examples/python-system-deps",
			want: true,
		},
		{
			name: "psycopg[binary]",
			path: "../../../examples/python-psycopg-binary",
			want: true,
		},
		{
			name: "psycopg (non-binary)",
			path: "../../../examples/python-latest-psycopg",
			want: false,
		},
		{
			name: "psycopg2 (django)",
			path: "../../../examples/python-django",
			want: false,
		},
		{
			name: "psycopg2 in workspace sub-package (non-binary)",
			path: "../../../examples/python-uv-workspace-postgres",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PythonProvider{}
			got := provider.usesBinaryPsycopg(ctx)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUsesPostgres(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "psycopg2-binary should not need apt packages",
			path: "../../../examples/python-system-deps",
			want: false,
		},
		{
			name: "psycopg[binary] should not need apt packages",
			path: "../../../examples/python-psycopg-binary",
			want: false,
		},
		{
			name: "psycopg (non-binary) needs apt packages",
			path: "../../../examples/python-latest-psycopg",
			want: true,
		},
		{
			name: "psycopg2 (django) needs apt packages",
			path: "../../../examples/python-django",
			want: true,
		},
		{
			name: "psycopg2 in workspace sub-package needs apt packages",
			path: "../../../examples/python-uv-workspace-postgres",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PythonProvider{}
			got := provider.usesPostgres(ctx)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUsesProductionPlaywrightDependency(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		dependency string
		want       bool
	}{
		{
			name:       "PEP 621 production dependency",
			path:       "../../../examples/python-playwright",
			dependency: "playwright",
			want:       true,
		},
		{
			// python-playwright declares pytest in dependency-groups.dev.
			name:       "PEP 735 development dependency",
			path:       "../../../examples/python-playwright",
			dependency: "pytest",
			want:       false,
		},
		{
			name:       "Poetry production dependency",
			path:       "../../../examples/python-poetry",
			dependency: "flask",
			want:       true,
		},
		{
			name:       "Pipenv production dependency",
			path:       "../../../examples/python-pipfile",
			dependency: "cowsay",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PythonProvider{}

			require.Equal(t, tt.want, provider.usesProductionDep(ctx, tt.dependency))
		})
	}
}

func TestPlaywrightInstallationIsOptIn(t *testing.T) {
	t.Run("does not install by default", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/python-playwright")
		provider := PythonProvider{}

		require.NoError(t, provider.Plan(ctx))
		require.False(t, stepHasExecCommand(ctx, "install", "playwright install --only-shell"))
		require.NotContains(t, ctx.Deploy.AptPackages, "libnss3")
	})

	t.Run("install when enabled", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/python-playwright")
		ctx.Env.Variables["RAILPACK_PYTHON_PLAYWRIGHT_INSTALL"] = "1"
		provider := PythonProvider{}

		require.NoError(t, provider.Plan(ctx))
		require.True(t, stepHasExecCommand(ctx, "install", "playwright install --only-shell"))
		require.Contains(t, ctx.Deploy.AptPackages, "libnss3")
	})
}

func TestDependencySpecMatches(t *testing.T) {
	require.True(t, dependencySpecMatches("Playwright[chromium]>=1.49", "playwright"))
	require.True(t, dependencySpecMatches("playwright @ https://example.com/playwright.whl", "playwright"))
	require.False(t, dependencySpecMatches("pytest-playwright>=0.7", "playwright"))
}

func stepHasExecCommand(ctx *generate.GenerateContext, stepName string, command string) bool {
	step := ctx.GetStepByName(stepName)
	if step == nil {
		return false
	}

	commandStep, ok := (*step).(*generate.CommandStepBuilder)
	if !ok {
		return false
	}

	for _, candidate := range commandStep.Commands {
		execCommand, ok := candidate.(plan.ExecCommand)
		if ok && execCommand.Cmd == command {
			return true
		}
	}

	return false
}
