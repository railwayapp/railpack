package python

import (
	"testing"

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

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple package name",
			input: "flask",
			want:  "flask",
		},
		{
			name:  "package with >= version",
			input: "flask>=3.0",
			want:  "flask",
		},
		{
			name:  "package with == version",
			input: "django==4.2.0",
			want:  "django",
		},
		{
			name:  "package with extras",
			input: "psycopg[binary]",
			want:  "psycopg[binary]",
		},
		{
			name:  "package with extras and version",
			input: "psycopg[binary]>=3.2",
			want:  "psycopg[binary]",
		},
		{
			name:  "package with ~= version",
			input: "requests~=2.31.0",
			want:  "requests",
		},
		{
			name:  "package with @ URL",
			input: "mypackage @ git+https://github.com/user/repo.git",
			want:  "mypackage",
		},
		{
			name:  "package with environment marker",
			input: "typing-extensions ; python_version < '3.8'",
			want:  "typing-extensions",
		},
		{
			name:  "package with spaces",
			input: "  numpy  >=  1.20  ",
			want:  "numpy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPackageName(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeDep(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercase simple package",
			input: "flask",
			want:  "flask",
		},
		{
			name:  "uppercase package",
			input: "Flask",
			want:  "flask",
		},
		{
			name:  "mixed case with version",
			input: "Django>=4.2",
			want:  "django",
		},
		{
			name:  "package with extras",
			input: "psycopg[binary]",
			want:  "psycopg[binary]",
		},
		{
			name:  "uppercase with extras and version",
			input: "Psycopg[Binary]>=3.2",
			want:  "psycopg[binary]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeDep(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestHasProductionDependency(t *testing.T) {
	tests := []struct {
		name string
		path string
		dep  string
		want bool
	}{
		{
			name: "pip - flask present",
			path: "../../../examples/python-flask",
			dep:  "flask",
			want: true,
		},
		{
			name: "pip - flask case insensitive",
			path: "../../../examples/python-flask",
			dep:  "Flask",
			want: true,
		},
		{
			name: "poetry - flask present",
			path: "../../../examples/python-poetry",
			dep:  "flask",
			want: true,
		},
		{
			name: "uv - flask present",
			path: "../../../examples/python-uv",
			dep:  "flask",
			want: true,
		},
		{
			name: "psycopg with extras",
			path: "../../../examples/python-psycopg-binary",
			dep:  "psycopg[binary]",
			want: true,
		},
		{
			name: "django present",
			path: "../../../examples/python-django",
			dep:  "django",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PythonProvider{}
			got := provider.HasProductionDependency(ctx, tt.dep)
			require.Equal(t, tt.want, got)
		})
	}
}
