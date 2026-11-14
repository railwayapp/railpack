package gleam

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGleamConfigIncludeSource(t *testing.T) {
	tests := []struct {
		name       string
		fileConfig string
		env        map[string]string
		want       bool
	}{
		{
			name: "environment truthy value remains supported",
			env:  map[string]string{"RAILPACK_GLEAM_INCLUDE_SOURCE": "1"},
			want: true,
		},
		{
			name:       "config enables source files",
			fileConfig: `{"gleam":{"includeSource":true}}`,
			want:       true,
		},
		{
			name:       "config false overrides truthy environment variable",
			fileConfig: `{"gleam":{"includeSource":false}}`,
			env:        map[string]string{"RAILPACK_GLEAM_INCLUDE_SOURCE": "TRUE"},
			want:       false,
		},
		{
			name:       "truthy environment variable applies without config value",
			fileConfig: `{"gleam":{}}`,
			env:        map[string]string{"RAILPACK_GLEAM_INCLUDE_SOURCE": "true"},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, "../../../examples/gleam")
			ctx.Env = app.NewEnvironment(&tt.env)

			fileConfig := config.EmptyConfig()
			if tt.fileConfig != "" {
				require.NoError(t, json.Unmarshal([]byte(tt.fileConfig), fileConfig))
			}
			ctx.Config = config.Merge(ctx.Config, fileConfig)

			provider := GleamProvider{}
			require.NoError(t, provider.Initialize(ctx))
			require.Equal(t, tt.want, provider.includeSource)
		})
	}

}

func TestGleamProviderConfigSchema(t *testing.T) {
	schema := (&GleamProvider{}).ConfigSchema()
	require.Equal(t, "object", schema.Type)
	propertySchema, ok := schema.Properties.Get("includeSource")
	require.True(t, ok)
	require.Equal(t, "boolean", propertySchema.Type)
}
