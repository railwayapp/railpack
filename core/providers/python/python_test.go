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
