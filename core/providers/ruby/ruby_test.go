package ruby

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
			path: "../../../examples/ruby-pip",
			want: true,
		},
		{
			name: "poetry",
			path: "../../../examples/ruby-poetry",
			want: true,
		},
		{
			name: "pdm",
			path: "../../../examples/ruby-pdm",
			want: true,
		},
		{
			name: "uv",
			path: "../../../examples/ruby-uv",
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
