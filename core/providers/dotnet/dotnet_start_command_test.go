package dotnet

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGetStartCommandUsesAssemblyName(t *testing.T) {
	dir := t.TempDir()
	projPath := filepath.Join(dir, "csharp-cli.csproj")
	require.NoError(t, os.WriteFile(projPath, []byte(`<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net6.0</TargetFramework>
    <AssemblyName>MyApp</AssemblyName>
  </PropertyGroup>
</Project>
`), 0o644))

	ctx := testingUtils.CreateGenerateContext(t, dir)
	provider := DotnetProvider{}

	require.Equal(t, "ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000} ./out/MyApp", provider.GetStartCommand(ctx))
}
