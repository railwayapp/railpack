package dotnet

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDotnet(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		detected bool
	}{
		{
			name:     "dotnet project",
			path:     "../../../examples/dotnet-cli",
			detected: true,
		},
		{
			name:     "non-dotnet project",
			path:     "../../../examples/node-npm",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := DotnetProvider{}

			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				err = provider.Plan(ctx)
				require.NoError(t, err)
			}
		})
	}
}
