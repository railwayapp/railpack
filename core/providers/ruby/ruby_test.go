package ruby

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "sinatra",
			path: "../../../examples/ruby-sinatra",
			want: true,
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
func TestGemfileRubyVersion(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "sinatra",
			path: "../../../examples/ruby-sinatra",
			want: "3.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := RubyProvider{}
			version := provider.gemfileRubyVersion(ctx)
			require.Equal(t, tt.want, version)
		})
	}
}
