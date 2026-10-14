package zig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testingUtils "github.com/railwayapp/railpack/core/testing"
)

func TestParseZonName(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "enum literal name",
			contents: ".{\n    .name = .my_app,\n    .version = \"0.0.0\",\n}",
			want:     "my_app",
		},
		{
			name:     "string name",
			contents: ".{\n    .name = \"my_app\",\n    .version = \"0.0.0\",\n}",
			want:     "my_app",
		},
		{
			name:     "commented example is skipped",
			contents: ".{\n    // .name = .from_the_comment,\n    .name = .real_name,\n}",
			want:     "real_name",
		},
		{
			name:     "trailing inline comment on name line",
			contents: ".{\n    .name = .my_app, // application entry point\n}",
			want:     "my_app",
		},
		{
			name:     "inline comment cannot shadow the real name",
			contents: ".{\n    .version = \"0.0.0\", // .name = .fake_name\n    .name = .real_name,\n}",
			want:     "real_name",
		},
		{
			name:     "slashes inside a string are not a comment",
			contents: ".{\n    .url = \"https://example.com/repo\", // .name = .fake_name\n    .name = .real_name,\n}",
			want:     "real_name",
		},
		{
			name:     "name-shaped text inside a string is ignored",
			contents: ".{\n    .url = \"https://example.com/.name = .fake_name\",\n    .name = .real_name,\n}",
			want:     "real_name",
		},
		{
			name:     "multiline string line is ignored",
			contents: ".{\n    .note =\n        \\\\ .name = .fake_name\n    ,\n    .name = .real_name,\n}",
			want:     "real_name",
		},
		{
			name:     "no name field",
			contents: ".{\n    .version = \"0.0.0\",\n}",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseZonName(tt.contents))
		})
	}
}

func TestZigDetectionAndStartCommand(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		detected bool
		startCmd string
	}{
		{
			name:     "zig",
			path:     "../../../examples/zig",
			detected: true,
			startCmd: "./zig-out/bin/zig_example",
		},
		{
			name:     "golang",
			path:     "../../../examples/go-mod",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := ZigProvider{}

			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if !detected {
				return
			}

			require.NoError(t, provider.Initialize(ctx))
			require.NoError(t, provider.Plan(ctx))
			assert.Equal(t, tt.startCmd, ctx.Deploy.StartCmd)
		})
	}
}
