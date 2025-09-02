package node

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageJson(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		packageJson := NewPackageJson()
		assert.Equal(t, packageJson.HasScript("start"), false)
		assert.Equal(t, packageJson.GetScript("build"), "")
		assert.Equal(t, packageJson.hasDependency("react"), false)
	})

	t.Run("full package.json", func(t *testing.T) {
		data := []byte(`{
			"name": "test-package",
			"version": "1.0.0",
			"packageManager": "yarn@4.7.0",
			"scripts": {
				"dev": "next dev",
				"build": "next build",
				"start": "next start"
			},
			"dependencies": {
				"next": "^14.0.0",
				"react": "^18.2.0"
			},
			"devDependencies": {
				"typescript": "^5.0.0",
				"@types/react": "^18.2.0"
			},
			"engines": {
				"node": ">=20 <21"
			},
			"main": "index.js"
		}`)

		var packageJson PackageJson
		err := json.Unmarshal(data, &packageJson)
		assert.NoError(t, err)

		// Test basic fields
		assert.Equal(t, "test-package", packageJson.Name)
		assert.Equal(t, "1.0.0", packageJson.Version)
		assert.Equal(t, "yarn@4.7.0", *packageJson.PackageManager)
		assert.Equal(t, "index.js", packageJson.Main)

		// Test scripts
		assert.True(t, packageJson.HasScript("start"))
		assert.Equal(t, "next start", packageJson.GetScript("start"))
		assert.Equal(t, "next build", packageJson.GetScript("build"))
		assert.Equal(t, "next dev", packageJson.GetScript("dev"))
		assert.False(t, packageJson.HasScript("test"))

		// Test dependencies
		assert.True(t, packageJson.hasDependency("next"))
		assert.True(t, packageJson.hasDependency("react"))
		assert.True(t, packageJson.hasDependency("typescript"))
		assert.True(t, packageJson.hasDependency("@types/react"))
		assert.False(t, packageJson.hasDependency("nonexistent"))

		// Test engines
		assert.Equal(t, ">=20 <21", packageJson.Engines["node"])
	})

	t.Run("workspaces array format", func(t *testing.T) {
		data := []byte(`{
			"name": "test-package",
			"workspaces": ["backend/**", "common", "frontend/**"]
		}`)

		var packageJson PackageJson
		err := json.Unmarshal(data, &packageJson)
		assert.NoError(t, err)
		assert.Equal(t, []string{"backend/**", "common", "frontend/**"}, packageJson.Workspaces)
	})

	t.Run("workspaces object format", func(t *testing.T) {
		data := []byte(`{
			"name": "test-package",
			"workspaces": {
				"packages": ["backend/**", "common", "frontend/**"]
			}
		}`)

		var packageJson PackageJson
		err := json.Unmarshal(data, &packageJson)
		assert.NoError(t, err)
		assert.Equal(t, []string{"backend/**", "common", "frontend/**"}, packageJson.Workspaces)
	})

	t.Run("empty workspaces", func(t *testing.T) {
		data := []byte(`{
			"name": "test-package",
			"workspaces": []
		}`)

		var packageJson PackageJson
		err := json.Unmarshal(data, &packageJson)
		assert.NoError(t, err)
		assert.Empty(t, packageJson.Workspaces)
	})

	t.Run("no workspaces field", func(t *testing.T) {
		data := []byte(`{
			"name": "test-package"
		}`)

		var packageJson PackageJson
		err := json.Unmarshal(data, &packageJson)
		assert.NoError(t, err)
		assert.Empty(t, packageJson.Workspaces)
	})

	t.Run("optional fields", func(t *testing.T) {
		data := []byte(`{
			"name": "test-package"
		}`)

		var packageJson PackageJson
		err := json.Unmarshal(data, &packageJson)
		assert.NoError(t, err)

		// Test optional fields are properly initialized
		assert.Nil(t, packageJson.PackageManager)
		assert.Empty(t, packageJson.Scripts)
		assert.Empty(t, packageJson.Dependencies)
		assert.Empty(t, packageJson.DevDependencies)
		assert.Empty(t, packageJson.Engines)
		assert.Empty(t, packageJson.Main)
	})
}

func TestGetPackageManagerInfo(t *testing.T) {
	tests := []struct {
		name           string
		packageManager string
		wantName       string
		wantVersion    string
	}{
		{
			name:           "valid package manager field",
			packageManager: "pnpm@8.15.4",
			wantName:       "pnpm",
			wantVersion:    "8.15.4",
		},
		{
			name:           "valid package manager field with yarn",
			packageManager: "yarn@4.1.0",
			wantName:       "yarn",
			wantVersion:    "4.1.0",
		},
		{
			name:           "valid package manager field with bun",
			packageManager: "bun@1.0.25",
			wantName:       "bun",
			wantVersion:    "1.0.25",
		},
		{
			name:           "valid package manager field with yarn and SHA",
			packageManager: "yarn@3.2.3+sha224.953c8233f7a92884eee2de69a1b92d1f2ec1655e66d08071ba9a02fa",
			wantName:       "yarn",
			wantVersion:    "3.2.3",
		},
		{
			name:           "empty package manager field",
			packageManager: "",
			wantName:       "",
			wantVersion:    "",
		},
		{
			name:           "invalid format - no version",
			packageManager: "pnpm",
			wantName:       "",
			wantVersion:    "",
		},
		{
			name:           "invalid format - multiple @ symbols",
			packageManager: "pnpm@8.15.4@extra",
			wantName:       "",
			wantVersion:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pm *string
			if tt.packageManager != "" {
				pm = &tt.packageManager
			}
			pkgJson := &PackageJson{
				PackageManager: pm,
			}
			gotName, gotVersion := pkgJson.GetPackageManagerInfo()
			if gotName != tt.wantName || gotVersion != tt.wantVersion {
				t.Errorf("GetPackageManagerInfo() = (%v, %v), want (%v, %v)",
					gotName, gotVersion, tt.wantName, tt.wantVersion)
			}
		})
	}
}
