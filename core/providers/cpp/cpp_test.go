package cpp

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestCppConfigVersions(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		fileConfig  string
		packageName string
		wantVersion string
	}{
		{
			name:        "CMake version",
			path:        "../../../examples/cpp-cmake",
			fileConfig:  `{"cpp":{"cmakeVersion":"3.30"}}`,
			packageName: "cmake",
			wantVersion: "3.30",
		},
		{
			name:        "Meson version",
			path:        "../../../examples/cpp-meson",
			fileConfig:  `{"cpp":{"mesonVersion":"1.5"}}`,
			packageName: "meson",
			wantVersion: "1.5",
		},
		{
			name:        "Ninja version",
			path:        "../../../examples/cpp-cmake",
			fileConfig:  `{"cpp":{"ninjaVersion":"1.12"}}`,
			packageName: "ninja",
			wantVersion: "1.12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			fileConfig := config.EmptyConfig()
			require.NoError(t, json.Unmarshal([]byte(tt.fileConfig), fileConfig))
			ctx.Config = config.Merge(ctx.Config, fileConfig)

			provider := CppProvider{}
			require.NoError(t, provider.Initialize(ctx))
			require.NoError(t, provider.Plan(ctx))
			require.Equal(t, tt.wantVersion, ctx.Resolver.Get(tt.packageName).Version)
		})
	}
}

func TestCppProviderConfigSchema(t *testing.T) {
	tests := []struct {
		property string
		wantType string
	}{
		{property: "cmakeVersion", wantType: "string"},
		{property: "mesonVersion", wantType: "string"},
		{property: "ninjaVersion", wantType: "string"},
	}

	schema := (&CppProvider{}).ConfigSchema()
	require.Equal(t, "object", schema.Type)
	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			propertySchema, ok := schema.Properties.Get(tt.property)
			require.True(t, ok)
			require.Equal(t, tt.wantType, propertySchema.Type)
		})
	}
}
