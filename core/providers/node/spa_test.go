package node

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestVite(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		isSPA         bool
		isVite        bool
		isAstro       bool
		isCRA         bool
		isAngular     bool
		isReactRouter bool
		outputDir     string
	}{
		{
			name:      "vite-react",
			path:      "../../../examples/node-vite-react",
			isSPA:     true,
			isVite:    true,
			outputDir: "dist",
		},
		{
			name:      "vite-svelte",
			path:      "../../../examples/node-vite-svelte",
			isSPA:     true,
			isVite:    true,
			outputDir: "theoutput",
		},
		{
			name:      "cra",
			path:      "../../../examples/node-cra",
			isSPA:     true,
			isCRA:     true,
			outputDir: "build",
		},
		{
			name:      "angular",
			path:      "../../../examples/node-angular",
			isSPA:     true,
			isAngular: true,
			outputDir: "dist/node-angular/browser",
		},
		{
			name:      "astro-static",
			path:      "../../../examples/node-astro",
			isSPA:     true,
			isAstro:   true,
			outputDir: "dist",
		},
		{
			name:      "astro-server",
			path:      "../../../examples/node-astro-server",
			isSPA:     false,
			isAstro:   true,
			outputDir: "dist",
		},
		{
			name:      "corepack",
			path:      "../../../examples/node-corepack",
			isSPA:     false,
			outputDir: "",
		},
		{
			name:      "golang",
			path:      "../../../examples/go-mod",
			isSPA:     false,
			outputDir: "",
		},
		{
			name:          "react-router",
			path:          "../../../examples/node-vite-react-router-spa",
			isSPA:         true,
			isVite:        true,
			isReactRouter: true,
			outputDir:     "build/client/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}

			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			if !detected {
				return
			}

			err = provider.Initialize(ctx)
			require.NoError(t, err)
			isSPA := provider.isSPA(ctx)
			require.Equal(t, tt.isSPA, isSPA)

			isVite := provider.isVite(ctx)
			require.Equal(t, tt.isVite, isVite)

			isAstro := provider.isAstro(ctx)
			require.Equal(t, tt.isAstro, isAstro)

			isReactRouter := provider.isReactRouter(ctx)
			require.Equal(t, tt.isReactRouter, isReactRouter)

			if tt.isSPA {
				require.Equal(t, tt.outputDir, provider.getOutputDirectory(ctx))
			}
		})
	}
}

func TestHasCustomStartCommand(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "vite-react",
			path: "../../../examples/node-vite-react",
			want: false,
		},
		{
			name: "angular",
			path: "../../../examples/node-angular",
			want: false,
		},
		{
			name: "astro static",
			path: "../../../examples/node-astro",
			want: false,
		},
		{
			name: "astro server",
			path: "../../../examples/node-astro-server",
			want: true,
		},
		{
			name: "npm",
			path: "../../../examples/node-npm",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}
			err := provider.Initialize(ctx)
			require.NoError(t, err)

			hasCustomStart := provider.hasCustomStartCommand(ctx)
			require.Equal(t, tt.want, hasCustomStart)
		})
	}
}
