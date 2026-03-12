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
