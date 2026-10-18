package dotnet

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

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
			name:     "last assembly name wins",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>First</AssemblyName>\n  </PropertyGroup>\n  <PropertyGroup>\n    <AssemblyName>Second</AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/Second",
		},
		{
			name:     "msbuild expression falls back to project file name",
			csproj:   "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net8.0</TargetFramework>\n    <AssemblyName>$(MSBuildProjectName).Server</AssemblyName>\n  </PropertyGroup>\n</Project>\n",
			startCmd: "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/my-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "my-project.csproj"), []byte(tt.csproj), 0o644))

			ctx := testingUtils.CreateGenerateContext(t, dir)
			provider := DotnetProvider{}
			require.NoError(t, provider.Plan(ctx))
			require.Equal(t, tt.startCmd, ctx.Deploy.StartCmd)
			require.Equal(t, tt.startCmd, provider.GetStartCommand(ctx))
		})
	}
}
