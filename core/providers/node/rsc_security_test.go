package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanVersionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"exact version", "19.0.0", "19.0.0"},
		{"caret version", "^19.0.0", "19.0.0"},
		{"tilde version", "~19.0.0", "19.0.0"},
		{"gte version", ">=19.0.0", "19.0.0"},
		{"lte version", "<=19.0.0", "19.0.0"},
		{"gt version", ">19.0.0", "19.0.0"},
		{"lt version", "<19.0.0", "19.0.0"},
		{"equal version", "=19.0.0", "19.0.0"},
		{"v prefix", "v19.0.0", "19.0.0"},
		{"range", "19.0.0 - 19.0.2", "19.0.0"},
		{"or range", "19.0.0 || 19.1.0", "19.0.0"},
		{"complex range", ">=19.0.0 <19.1.0", "19.0.0"},
		{"with spaces", "  ^19.0.0  ", "19.0.0"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanVersionString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFixedVersionForRSC(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		wantFixed      string
		wantVulnerable bool
	}{
		// Vulnerable 19.0.x versions
		{"19.0.0 is vulnerable", "19.0.0", "19.0.3", true},
		{"19.0.1 is vulnerable", "19.0.1", "19.0.3", true},
		{"19.0.2 is vulnerable", "19.0.2", "19.0.3", true},
		{"^19.0.0 is vulnerable", "^19.0.0", "19.0.3", true},
		{"~19.0.1 is vulnerable", "~19.0.1", "19.0.3", true},

		// Fixed 19.0.x versions
		{"19.0.3 is safe", "19.0.3", "", false},
		{"19.0.4 is safe", "19.0.4", "", false},

		// Vulnerable 19.1.x versions
		{"19.1.0 is vulnerable", "19.1.0", "19.1.4", true},
		{"19.1.1 is vulnerable", "19.1.1", "19.1.4", true},
		{"19.1.2 is vulnerable", "19.1.2", "19.1.4", true},
		{"19.1.3 is vulnerable", "19.1.3", "19.1.4", true},
		{"^19.1.0 is vulnerable", "^19.1.0", "19.1.4", true},

		// Fixed 19.1.x versions
		{"19.1.4 is safe", "19.1.4", "", false},
		{"19.1.5 is safe", "19.1.5", "", false},

		// Vulnerable 19.2.x versions
		{"19.2.0 is vulnerable", "19.2.0", "19.2.3", true},
		{"19.2.1 is vulnerable", "19.2.1", "19.2.3", true},
		{"19.2.2 is vulnerable", "19.2.2", "19.2.3", true},
		{"^19.2.0 is vulnerable", "^19.2.0", "19.2.3", true},

		// Fixed 19.2.x versions
		{"19.2.3 is safe", "19.2.3", "", false},
		{"19.2.4 is safe", "19.2.4", "", false},

		// Non-19.x versions (not affected)
		{"18.0.0 is not affected", "18.0.0", "", false},
		{"18.2.0 is not affected", "^18.2.0", "", false},
		{"20.0.0 is not affected", "20.0.0", "", false},

		// Future 19.x versions (assume safe)
		{"19.3.0 is assumed safe", "19.3.0", "", false},
		{"19.4.0 is assumed safe", "19.4.0", "", false},

		// Invalid versions
		{"empty version", "", "", false},
		{"invalid version", "invalid", "", false},
		{"latest", "latest", "", false},
		{"*", "*", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFixedVersionForRSC(tt.version)
			if tt.wantVulnerable {
				assert.Equal(t, tt.wantFixed, result, "expected vulnerable version to return fixed version")
			} else {
				assert.Empty(t, result, "expected safe version to return empty string")
			}
		})
	}
}

func TestPackageJsonGetDependencyVersion(t *testing.T) {
	t.Run("gets version from dependencies", func(t *testing.T) {
		pkgJson := &PackageJson{
			Dependencies: map[string]string{
				"react-server-dom-webpack": "19.0.0",
				"react":                    "^19.0.0",
			},
		}
		assert.Equal(t, "19.0.0", pkgJson.getDependencyVersion("react-server-dom-webpack"))
		assert.Equal(t, "^19.0.0", pkgJson.getDependencyVersion("react"))
	})

	t.Run("gets version from devDependencies", func(t *testing.T) {
		pkgJson := &PackageJson{
			DevDependencies: map[string]string{
				"react-server-dom-webpack": "19.1.0",
			},
		}
		assert.Equal(t, "19.1.0", pkgJson.getDependencyVersion("react-server-dom-webpack"))
	})

	t.Run("returns empty for missing dependency", func(t *testing.T) {
		pkgJson := &PackageJson{
			Dependencies: map[string]string{
				"react": "^19.0.0",
			},
		}
		assert.Empty(t, pkgJson.getDependencyVersion("react-server-dom-webpack"))
	})

	t.Run("prefers dependencies over devDependencies", func(t *testing.T) {
		pkgJson := &PackageJson{
			Dependencies: map[string]string{
				"react-server-dom-webpack": "19.0.0",
			},
			DevDependencies: map[string]string{
				"react-server-dom-webpack": "19.1.0",
			},
		}
		assert.Equal(t, "19.0.0", pkgJson.getDependencyVersion("react-server-dom-webpack"))
	})

	t.Run("handles nil maps", func(t *testing.T) {
		pkgJson := &PackageJson{}
		assert.Empty(t, pkgJson.getDependencyVersion("react-server-dom-webpack"))
	})
}

func TestVulnerableRSCPackages(t *testing.T) {
	// Ensure we're checking all vulnerable RSC packages
	expectedPackages := []string{
		"react-server-dom-webpack",
		"react-server-dom-parcel",
		"react-server-dom-turbopack",
	}

	assert.Equal(t, expectedPackages, vulnerableRSCPackages)
}

func TestRSCAffectedFrameworks(t *testing.T) {
	// Ensure we're checking all affected frameworks
	expectedFrameworks := []string{
		"waku",
		"@parcel/rsc",
		"@vitejs/plugin-rsc",
		"rwsdk",
	}

	assert.Equal(t, expectedFrameworks, rscAffectedFrameworks)
}

func TestRSCFixedVersions(t *testing.T) {
	// Ensure fixed versions are correctly defined
	assert.Equal(t, "19.0.3", rscFixedVersions["19.0"])
	assert.Equal(t, "19.1.4", rscFixedVersions["19.1"])
	assert.Equal(t, "19.2.3", rscFixedVersions["19.2"])
}

func TestGetFixedVersionForNextJS(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		wantFixed      string
		wantVulnerable bool
	}{
		// Next.js 13.x versions (13.3+ affected)
		{"13.0.0 is not affected", "13.0.0", "", false},
		{"13.1.0 is not affected", "13.1.0", "", false},
		{"13.2.0 is not affected", "13.2.0", "", false},
		{"13.3.0 is vulnerable", "13.3.0", "14.2.35", true},
		{"13.4.0 is vulnerable", "13.4.0", "14.2.35", true},
		{"13.5.0 is vulnerable", "13.5.0", "14.2.35", true},
		{"^13.4.0 is vulnerable", "^13.4.0", "14.2.35", true},

		// Next.js 14.x versions
		{"14.0.0 is vulnerable", "14.0.0", "14.2.35", true},
		{"14.1.0 is vulnerable", "14.1.0", "14.2.35", true},
		{"14.2.0 is vulnerable", "14.2.0", "14.2.35", true},
		{"14.2.34 is vulnerable", "14.2.34", "14.2.35", true},
		{"^14.0.0 is vulnerable", "^14.0.0", "14.2.35", true},
		{"14.2.35 is safe", "14.2.35", "", false},
		{"14.2.36 is safe", "14.2.36", "", false},

		// Next.js 15.x versions
		{"15.0.0 is vulnerable", "15.0.0", "15.0.7", true},
		{"15.0.6 is vulnerable", "15.0.6", "15.0.7", true},
		{"15.0.7 is safe", "15.0.7", "", false},
		{"15.1.0 is vulnerable", "15.1.0", "15.1.11", true},
		{"15.1.10 is vulnerable", "15.1.10", "15.1.11", true},
		{"15.1.11 is safe", "15.1.11", "", false},
		{"15.2.0 is vulnerable", "15.2.0", "15.2.8", true},
		{"15.2.7 is vulnerable", "15.2.7", "15.2.8", true},
		{"15.2.8 is safe", "15.2.8", "", false},
		{"15.3.0 is vulnerable", "15.3.0", "15.3.8", true},
		{"15.3.7 is vulnerable", "15.3.7", "15.3.8", true},
		{"15.3.8 is safe", "15.3.8", "", false},
		{"15.4.0 is vulnerable", "15.4.0", "15.4.10", true},
		{"15.4.9 is vulnerable", "15.4.9", "15.4.10", true},
		{"15.4.10 is safe", "15.4.10", "", false},
		{"15.5.0 is vulnerable", "15.5.0", "15.5.9", true},
		{"15.5.8 is vulnerable", "15.5.8", "15.5.9", true},
		{"15.5.9 is safe", "15.5.9", "", false},

		// Next.js 16.x versions
		{"16.0.0 is vulnerable", "16.0.0", "16.0.10", true},
		{"16.0.9 is vulnerable", "16.0.9", "16.0.10", true},
		{"16.0.10 is safe", "16.0.10", "", false},

		// Future versions (assume safe)
		{"15.6.0 is assumed safe", "15.6.0", "", false},
		{"16.1.0 is assumed safe", "16.1.0", "", false},
		{"17.0.0 is assumed safe", "17.0.0", "", false},

		// Old versions not affected
		{"12.0.0 is not affected", "12.0.0", "", false},
		{"11.0.0 is not affected", "11.0.0", "", false},

		// Invalid versions
		{"empty version", "", "", false},
		{"invalid version", "invalid", "", false},
		{"latest", "latest", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFixedVersionForNextJS(tt.version)
			if tt.wantVulnerable {
				assert.Equal(t, tt.wantFixed, result, "expected vulnerable version to return fixed version")
			} else {
				assert.Empty(t, result, "expected safe version to return empty string")
			}
		})
	}
}

func TestNextJSFixedVersions(t *testing.T) {
	// Ensure fixed versions are correctly defined for Next.js
	assert.Equal(t, "14.2.35", nextjsFixedVersions["13.3"])
	assert.Equal(t, "14.2.35", nextjsFixedVersions["13.4"])
	assert.Equal(t, "14.2.35", nextjsFixedVersions["13.5"])
	assert.Equal(t, "14.2.35", nextjsFixedVersions["14.0"])
	assert.Equal(t, "14.2.35", nextjsFixedVersions["14.1"])
	assert.Equal(t, "14.2.35", nextjsFixedVersions["14.2"])
	assert.Equal(t, "15.0.7", nextjsFixedVersions["15.0"])
	assert.Equal(t, "15.1.11", nextjsFixedVersions["15.1"])
	assert.Equal(t, "15.2.8", nextjsFixedVersions["15.2"])
	assert.Equal(t, "15.3.8", nextjsFixedVersions["15.3"])
	assert.Equal(t, "15.4.10", nextjsFixedVersions["15.4"])
	assert.Equal(t, "15.5.9", nextjsFixedVersions["15.5"])
	assert.Equal(t, "16.0.10", nextjsFixedVersions["16.0"])
}

func TestGetFixedVersionForReactRouter(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		wantFixed      string
		wantVulnerable bool
	}{
		// React Router 6.x is not affected (no RSC support)
		{"6.0.0 is not affected", "6.0.0", "", false},
		{"6.28.0 is not affected", "6.28.0", "", false},

		// React Router 7.x versions
		{"7.0.0 is vulnerable", "7.0.0", "7.0.3", true},
		{"7.0.2 is vulnerable", "7.0.2", "7.0.3", true},
		{"7.0.3 is safe", "7.0.3", "", false},
		{"7.1.0 is vulnerable", "7.1.0", "7.1.6", true},
		{"7.1.5 is vulnerable", "7.1.5", "7.1.6", true},
		{"7.1.6 is safe", "7.1.6", "", false},
		{"7.2.0 is vulnerable", "7.2.0", "7.2.3", true},
		{"7.2.2 is vulnerable", "7.2.2", "7.2.3", true},
		{"7.2.3 is safe", "7.2.3", "", false},
		{"7.3.0 is vulnerable", "7.3.0", "7.3.3", true},
		{"7.3.2 is vulnerable", "7.3.2", "7.3.3", true},
		{"7.3.3 is safe", "7.3.3", "", false},
		{"7.4.0 is vulnerable", "7.4.0", "7.4.2", true},
		{"7.4.1 is vulnerable", "7.4.1", "7.4.2", true},
		{"7.4.2 is safe", "7.4.2", "", false},
		{"7.5.0 is vulnerable", "7.5.0", "7.5.3", true},
		{"7.5.2 is vulnerable", "7.5.2", "7.5.3", true},
		{"7.5.3 is safe", "7.5.3", "", false},
		{"7.6.0 is vulnerable", "7.6.0", "7.6.3", true},
		{"7.6.2 is vulnerable", "7.6.2", "7.6.3", true},
		{"7.6.3 is safe", "7.6.3", "", false},
		{"^7.0.0 is vulnerable", "^7.0.0", "7.0.3", true},

		// Future versions (assume safe)
		{"7.7.0 is assumed safe", "7.7.0", "", false},
		{"8.0.0 is assumed safe", "8.0.0", "", false},

		// Invalid versions
		{"empty version", "", "", false},
		{"invalid version", "invalid", "", false},
		{"latest", "latest", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFixedVersionForReactRouter(tt.version)
			if tt.wantVulnerable {
				assert.Equal(t, tt.wantFixed, result, "expected vulnerable version to return fixed version")
			} else {
				assert.Empty(t, result, "expected safe version to return empty string")
			}
		})
	}
}

func TestReactRouterFixedVersions(t *testing.T) {
	// Ensure fixed versions are correctly defined for React Router
	assert.Equal(t, "7.0.3", reactRouterFixedVersions["7.0"])
	assert.Equal(t, "7.1.6", reactRouterFixedVersions["7.1"])
	assert.Equal(t, "7.2.3", reactRouterFixedVersions["7.2"])
	assert.Equal(t, "7.3.3", reactRouterFixedVersions["7.3"])
	assert.Equal(t, "7.4.2", reactRouterFixedVersions["7.4"])
	assert.Equal(t, "7.5.3", reactRouterFixedVersions["7.5"])
	assert.Equal(t, "7.6.3", reactRouterFixedVersions["7.6"])
}

func TestGetOverrideCommands(t *testing.T) {
	tests := []struct {
		name           string
		packageManager PackageManager
		expectedField  string
	}{
		{"npm uses overrides", PackageManagerNpm, "overrides"},
		{"pnpm uses pnpm.overrides", PackageManagerPnpm, "pnpm.overrides"},
		{"yarn1 uses resolutions", PackageManagerYarn1, "resolutions"},
		{"yarnberry uses resolutions", PackageManagerYarnBerry, "resolutions"},
		{"bun uses overrides", PackageManagerBun, "overrides"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &NodeProvider{packageManager: tt.packageManager}
			overrides := map[string]string{
				"next": "15.0.7",
			}

			commands := provider.getOverrideCommands(nil, overrides)

			assert.Len(t, commands, 1)
			assert.Contains(t, commands[0], "node -e")
			assert.Contains(t, commands[0], tt.expectedField)
			assert.Contains(t, commands[0], `"next": "15.0.7"`)
		})
	}
}

func TestGetOverrideCommandsWithScopedPackages(t *testing.T) {
	provider := &NodeProvider{packageManager: PackageManagerNpm}
	overrides := map[string]string{
		"@react-router/dev":        "7.5.3",
		"react-server-dom-webpack": "19.2.3",
	}

	commands := provider.getOverrideCommands(nil, overrides)

	assert.Len(t, commands, 1)
	// Verify scoped package is properly quoted in JSON
	assert.Contains(t, commands[0], `"@react-router/dev": "7.5.3"`)
	assert.Contains(t, commands[0], `"react-server-dom-webpack": "19.2.3"`)
}

func TestGetOverrideCommandsEmpty(t *testing.T) {
	provider := &NodeProvider{packageManager: PackageManagerNpm}

	commands := provider.getOverrideCommands(nil, map[string]string{})
	assert.Nil(t, commands)

	commands = provider.getOverrideCommands(nil, nil)
	assert.Nil(t, commands)
}
