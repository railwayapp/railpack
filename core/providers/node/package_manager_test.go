package node

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/resolver"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGetPackageManagerPackages_PnpmLockfileVersion(t *testing.T) {
	tests := []struct {
		name        string
		lockfile    string
		wantVersion string
		wantSource  string
	}{
		{
			name:        "lockfile 5.3",
			lockfile:    "lockfileVersion: 5.3\n",
			wantVersion: "6",
			wantSource:  "pnpm-lock.yaml",
		},
		{
			name:        "lockfile 5.4",
			lockfile:    "lockfileVersion: 5.4\n",
			wantVersion: "7",
			wantSource:  "pnpm-lock.yaml",
		},
		{
			name:        "lockfile 6.0 quoted",
			lockfile:    "lockfileVersion: '6.0'\n",
			wantVersion: "8",
			wantSource:  "pnpm-lock.yaml",
		},
		{
			name:        "lockfile 6.0 unquoted",
			lockfile:    "lockfileVersion: 6.0\n",
			wantVersion: "8",
			wantSource:  "pnpm-lock.yaml",
		},
		{
			name:        "lockfile 6.1",
			lockfile:    "lockfileVersion: '6.1'\n",
			wantVersion: "8",
			wantSource:  "pnpm-lock.yaml",
		},
		{
			name:        "lockfile 9.0 quoted",
			lockfile:    "lockfileVersion: '9.0'\n",
			wantVersion: "9",
			wantSource:  "pnpm-lock.yaml",
		},
		{
			name:        "lockfile 9.0 unquoted",
			lockfile:    "lockfileVersion: 9.0\n",
			wantVersion: "9",
			wantSource:  "pnpm-lock.yaml",
		},
		{
			name:        "unknown lockfile keeps default",
			lockfile:    "lockfileVersion: '99.0'\n",
			wantVersion: DEFAULT_PNPM_VERSION,
			wantSource:  resolver.DefaultSource,
		},
		{
			name:        "no lockfile keeps default",
			lockfile:    "",
			wantVersion: DEFAULT_PNPM_VERSION,
			wantSource:  resolver.DefaultSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0o644))
			if tt.lockfile != "" {
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte(tt.lockfile), 0o644))
			}

			ctx := testingUtils.CreateGenerateContext(t, tmpDir)
			miseStep := ctx.NewMiseStepBuilder("test")

			PackageManagerPnpm.GetPackageManagerPackages(ctx, NewPackageJson(), miseStep)

			pnpm := ctx.Resolver.Get("pnpm")
			require.NotNil(t, pnpm)
			require.Equal(t, tt.wantVersion, pnpm.Version)
			require.Equal(t, tt.wantSource, pnpm.Source)
		})
	}
}

func TestGetPackageManagerPackages_PnpmVersionPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte("lockfileVersion: '6.0'\n"), 0o644))

	t.Run("engines override lockfile", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		miseStep := ctx.NewMiseStepBuilder("test")
		packageJson := NewPackageJson()
		packageJson.Engines["pnpm"] = "10"

		PackageManagerPnpm.GetPackageManagerPackages(ctx, packageJson, miseStep)

		pnpm := ctx.Resolver.Get("pnpm")
		require.NotNil(t, pnpm)
		require.Equal(t, "10", pnpm.Version)
		require.Equal(t, "package.json > engines > pnpm", pnpm.Source)
	})

	t.Run("packageManager overrides lockfile", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		miseStep := ctx.NewMiseStepBuilder("test")
		pm := "pnpm@10.4.1"
		packageJson := &PackageJson{PackageManager: &pm}

		PackageManagerPnpm.GetPackageManagerPackages(ctx, packageJson, miseStep)

		pnpm := ctx.Resolver.Get("pnpm")
		require.NotNil(t, pnpm)
		require.Equal(t, "10.4.1", pnpm.Version)
		require.Equal(t, "package.json > packageManager", pnpm.Source)
		require.True(t, pnpm.SkipMiseInstall)
	})
}
