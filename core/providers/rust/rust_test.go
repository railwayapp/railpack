package rust

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestRust(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		detected    bool
		rustVersion string
	}{
		{
			name:        "rust system deps",
			path:        "../../../examples/rust-system-deps",
			detected:    true,
			rustVersion: "1.85.1",
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
			provider := RustProvider{}
			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				err = provider.Plan(ctx)
				require.NoError(t, err)

				if tt.rustVersion != "" {
					rustVersion := ctx.Resolver.Get("rust")
					require.Equal(t, tt.rustVersion, rustVersion.Version)
				}
			}
		})
	}
}
