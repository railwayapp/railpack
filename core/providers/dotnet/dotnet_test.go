package dotnet

import (
	"os"
	"path/filepath"
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

				require.Equal(t, "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/csharp-cli.App", ctx.Deploy.StartCmd)
			}
		})
	}
}

func TestGetStartCommandUsesAssemblyName(t *testing.T) {
	tests := []struct {
		name     string
		csproj   string
		startCmd string
	}{
		{
			name:     "project file name by default",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <OutputType>Exe</OutputType>\n    <TargetFramework>net8.0</TargetFramework>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/my-project",
		},
		{
			name:     "assembly name",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <OutputType>Exe</OutputType>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>MyApp</AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/MyApp",
		},
		{
			name:     "assembly name with surrounding whitespace",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <OutputType>Exe</OutputType>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>\n      MyApp\n    </AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/MyApp",
		},
		{
			name:     "last unconditional assembly name wins",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>First</AssemblyName>\n  </PropertyGroup>\n  <PropertyGroup>\n    <AssemblyName>Second</AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/Second",
		},
		{
			name:     "conditional property group and assembly name are ignored",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>MyApp</AssemblyName>\n    <AssemblyName Condition=\"'$(Configuration)' == 'Debug'\">MyApp.InlineDebug</AssemblyName>\n  </PropertyGroup>\n  <PropertyGroup Condition=\"'$(Configuration)' == 'Debug'\">\n    <AssemblyName>MyApp.Debug</AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/MyApp",
		},
		{
			name:     "MSBuildProjectName expression is resolved",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>$(MSBuildProjectName).Server</AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/my-project.Server",
		},
		{
			name:     "unknown msbuild expression falls back to project file name",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>$(CustomPrefix).Server</AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/my-project",
		},
		{
			name:     "unknown msbuild expression does not overwrite earlier valid assembly name",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>MyApp</AssemblyName>\n  </PropertyGroup>\n  <PropertyGroup>\n    <AssemblyName>$(CustomPrefix).Server</AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/MyApp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "my-project.csproj"), []byte(tt.csproj), 0o644))

			ctx := testingUtils.CreateGenerateContext(t, dir)
			provider := DotnetProvider{}
			require.Equal(t, tt.startCmd, provider.GetStartCommand(ctx))
		})
	}
}

func TestDotnetVersionResolution(t *testing.T) {
	tests := []struct {
		name            string
		csproj          string
		globalJSON      string
		envVersion      string
		expectedVersion string
	}{
		{
			name:            "platform-suffixed TargetFramework",
			csproj:          "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net7.0-linux</TargetFramework>\n  </PropertyGroup>\n</Project>\n",
			expectedVersion: "7.0",
		},
		{
			name:            "plural TargetFrameworks takes first",
			csproj:          "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFrameworks>net6.0;net8.0</TargetFrameworks>\n  </PropertyGroup>\n</Project>\n",
			expectedVersion: "6.0",
		},
		{
			name:            "global.json overrides csproj",
			csproj:          "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net6.0</TargetFramework>\n  </PropertyGroup>\n</Project>\n",
			globalJSON:      `{"sdk": {"version": "8.0.404"}}`,
			expectedVersion: "8.0.404",
		},
		{
			name:            "RAILPACK_DOTNET_VERSION overrides global.json and csproj",
			csproj:          "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net6.0</TargetFramework>\n  </PropertyGroup>\n</Project>\n",
			globalJSON:      `{"sdk": {"version": "8.0.404"}}`,
			envVersion:      "9.0",
			expectedVersion: "9.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "my-project.csproj"), []byte(tt.csproj), 0o644))
			if tt.globalJSON != "" {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "global.json"), []byte(tt.globalJSON), 0o644))
			}

			ctx := testingUtils.CreateGenerateContext(t, dir)
			if tt.envVersion != "" {
				ctx.Env.SetVariable("RAILPACK_DOTNET_VERSION", tt.envVersion)
			}

			provider := DotnetProvider{}
			miseStep := ctx.GetMiseStepBuilder()
			provider.InstallMisePackages(ctx, miseStep)

			pkg := miseStep.Resolver.Get("dotnet")
			require.NotNil(t, pkg)
			require.Equal(t, tt.expectedVersion, pkg.Version)
		})
	}
}
