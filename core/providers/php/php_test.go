package php

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestPhpProvider(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		isPhp     bool
		framework string
	}{
		{
			name:      "vanilla php with index.php",
			path:      "../../../examples/php-vanilla",
			isPhp:     true,
			framework: "",
		},
		{
			name:      "laravel project with composer.json",
			path:      "../../../examples/php-laravel-12-react",
			isPhp:     true,
			framework: "laravel",
		},
		{
			name:      "symfony project",
			path:      "../../../examples/php-symfony",
			isPhp:     true,
			framework: "symfony",
		},
		{
			name:      "non-php project",
			path:      "../../../examples/node-npm",
			isPhp:     false,
			framework: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PhpProvider{}

			isPhp, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.isPhp, isPhp)

			require.Equal(t, tt.framework, provider.getFramework(ctx))
		})
	}
}

// PHP installs build and deploy apt packages on one image, so "..." is removed and the lists are merged.
func TestPhpCombinesBuildAndDeployAptPackages(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/php-vanilla")
	ctx.Config.BuildAptPackages = []string{"...", "curl", "git"}
	ctx.Config.Deploy.AptPackages = []string{"...", "curl", "jq"}

	provider := PhpProvider{}
	require.NoError(t, provider.Initialize(ctx))
	require.NoError(t, provider.Plan(ctx))

	var packages []string
	for _, step := range ctx.Steps {
		imageStep, ok := step.(*generate.ImageStepBuilder)
		if !ok {
			continue
		}
		packages = imageStep.AptPackages
	}

	require.Equal(t, []string{"git", "zip", "unzip", "ca-certificates", "curl", "jq"}, packages)
}

func TestPhpConfigRootDir(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		configJSON  string
		envRootDir  string
		wantRootDir string
	}{
		{
			name:        "vanilla default",
			path:        "../../../examples/php-vanilla",
			wantRootDir: "/app",
		},
		{
			name:        "framework default",
			path:        "../../../examples/php-laravel-12-react",
			wantRootDir: "/app/public",
		},
		{
			name:        "legacy environment variable",
			path:        "../../../examples/php-vanilla",
			envRootDir:  "/legacy/public",
			wantRootDir: "/legacy/public",
		},
		{
			name:        "config file overrides environment variable",
			path:        "../../../examples/php-vanilla",
			configJSON:  `{"php":{"rootDir":"/configured/public"}}`,
			envRootDir:  "/legacy/public",
			wantRootDir: "/configured/public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			if tt.configJSON != "" {
				fileConfig := config.EmptyConfig()
				require.NoError(t, json.Unmarshal([]byte(tt.configJSON), fileConfig))
				ctx.Config = config.Merge(ctx.Config, fileConfig)
			}
			if tt.envRootDir != "" {
				ctx.Env.Variables["RAILPACK_PHP_ROOT_DIR"] = tt.envRootDir
			}

			provider := PhpProvider{}
			require.NoError(t, provider.Initialize(ctx))
			files, err := provider.getConfigFiles(ctx)
			require.NoError(t, err)
			require.Contains(t, files.Caddyfile.Contents, "root * "+tt.wantRootDir)
		})
	}
}

func TestPhpConfigSchema(t *testing.T) {
	schema := (&PhpProvider{}).ConfigSchema()
	versionSchema, ok := schema.Properties.Get("version")
	require.True(t, ok)
	require.Equal(t, "string", versionSchema.Type)
	rootDirSchema, ok := schema.Properties.Get("rootDir")
	require.True(t, ok)
	require.Equal(t, "string", rootDirSchema.Type)
	require.True(t, strings.Contains(rootDirSchema.Description, "document root"))
}

func TestPhpConfigVersion(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		configJSON    string
		envVersion    string
		wantVersion   string
		wantVersionSrc string
	}{
		{
			name:           "composer constraint remains supported",
			path:           "../../../examples/php-laravel-12-react",
			wantVersionSrc: "composer.json > require > php",
		},
		{
			name:           "legacy environment variable",
			path:           "../../../examples/php-vanilla",
			envVersion:     "8.3",
			wantVersion:    "8.3",
			wantVersionSrc: "RAILPACK_PHP_VERSION environment variable",
		},
		{
			name:           "config overrides environment variable",
			path:           "../../../examples/php-vanilla",
			configJSON:     `{"php":{"version":"8.2"}}`,
			envVersion:     "8.3",
			wantVersion:    "8.2",
			wantVersionSrc: "railpack config > php > version",
		},
		{
			name:           "config overrides composer constraint",
			path:           "../../../examples/php-laravel-12-react",
			configJSON:     `{"php":{"version":"8.2"}}`,
			wantVersion:    "8.2",
			wantVersionSrc: "railpack config > php > version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			if tt.configJSON != "" {
				fileConfig := config.EmptyConfig()
				require.NoError(t, json.Unmarshal([]byte(tt.configJSON), fileConfig))
				ctx.Config = config.Merge(ctx.Config, fileConfig)
			}
			if tt.envVersion != "" {
				ctx.Env.Variables["RAILPACK_PHP_VERSION"] = tt.envVersion
			}

			provider := PhpProvider{}
			require.NoError(t, provider.Initialize(ctx))
			_, err := provider.phpImagePackage(ctx)
			require.NoError(t, err)

			version := ctx.Resolver.Get("php")
			if tt.wantVersion != "" {
				require.Equal(t, tt.wantVersion, version.Version)
			}
			require.Equal(t, tt.wantVersionSrc, version.Source)
		})
	}
}
