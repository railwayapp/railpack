package python

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name string
		path string
		want PackageManager
	}{
		{
			name: "pip",
			path: "../../../examples/python-pip",
			want: PackageManagerPip,
		},
		{
			name: "poetry",
			path: "../../../examples/python-poetry",
			want: PackageManagerPoetry,
		},
		{
			name: "pdm",
			path: "../../../examples/python-pdm",
			want: PackageManagerPdm,
		},
		{
			name: "uv",
			path: "../../../examples/python-uv",
			want: PackageManagerUv,
		},
		{
			name: "pipfile",
			path: "../../../examples/python-pipfile",
			want: PackageManagerPipenv,
		},
		{
			name: "bot.py only",
			path: "../../../examples/python-bot-only",
			want: PackageManagerNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PythonProvider{}
			got := provider.detectPackageManager(ctx)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRequirementsNeedLocalFiles(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "relative file scheme wheel",
			content: "dummy @ file:_vendor/dummy-1.0.0-py3-none-any.whl\n",
			want:    true,
		},
		{
			name:    "relative file scheme wheel with hash",
			content: "dummy @ file:_vendor/dummy.whl --hash=sha256:abc\n",
			want:    true,
		},
		{
			name:    "file double slash scheme",
			content: "pkg @ file://./vendor/pkg.whl\n",
			want:    true,
		},
		{
			name:    "local git repository",
			content: "pkg @ git+file:///app/vendor/pkg\n",
			want:    true,
		},
		{
			name:    "bare local git repository",
			content: "git+file:///app/vendor/pkg#egg=pkg\n",
			want:    true,
		},
		{
			name:    "local subversion repository",
			content: "pkg @ svn+file:///app/vendor/pkg\n",
			want:    true,
		},
		{
			name:    "local mercurial repository",
			content: "pkg @ hg+file:///app/vendor/pkg\n",
			want:    true,
		},
		{
			name:    "local bazaar repository",
			content: "pkg @ bzr+file:///app/vendor/pkg\n",
			want:    true,
		},
		{
			name:    "bare relative wheel path",
			content: "_vendor/pkg-1.0.whl\n",
			want:    true,
		},
		{
			name:    "short editable flag",
			content: "-e ./local-package\n",
			want:    true,
		},
		{
			name:    "long editable flag",
			content: "--editable .\n",
			want:    true,
		},
		{
			name:    "PEP 508 local directory path",
			content: "pkg @ ./local-lib\n",
			want:    true,
		},
		{
			name:    "local find-links directory",
			content: "-f _vendor\nrequests==2.31.0\n",
			want:    true,
		},
		{
			name:    "remote find-links URL",
			content: "-f https://download.pytorch.org/whl/cpu\n",
			want:    false,
		},
		{
			name:    "standard PyPI pin",
			content: "requests==2.31.0\n",
			want:    false,
		},
		{
			name:    "remote HTTPS wheel",
			content: "pkg @ https://example.com/pkg.whl\n",
			want:    false,
		},
		{
			name:    "remote git requirement",
			content: "pkg @ git+https://github.com/org/repo.git\n",
			want:    false,
		},
		{
			name:    "comment mentioning file scheme",
			content: "# vendored from file:_vendor/foo.whl\nrequests==2.31.0\n",
			want:    false,
		},
		{
			name:    "git requirement from python-pip example",
			content: "fastapi-ipware @ git+https://github.com/iloveitaly/fastapi-ipware.git\n",
			want:    false,
		},
		{
			name:    "compact editable local file URL",
			content: "-efile:///app/vendor/pkg\n",
			want:    true,
		},
		{
			name:    "compact editable local git file URL",
			content: "-egit+file:///app/vendor/pkg#egg=pkg\n",
			want:    true,
		},
		{
			name:    "compact local find-links",
			content: "-ffile:///app/vendor\n",
			want:    true,
		},
		{
			name:    "spaced editable local file URL",
			content: "-e file:///app/vendor/pkg\n",
			want:    true,
		},
		{
			name:    "spaced local find-links file URL",
			content: "-f file:///app/vendor\n",
			want:    true,
		},
		{
			name:    "compact remote find-links URL",
			content: "-fhttps://example.com/wheels\n",
			want:    false,
		},
		{
			name:    "remote git requirement spec",
			content: "pkg @ git+https://example.com/pkg.git\n",
			want:    false,
		},
		{
			name:    "comment with compact local option",
			content: "# -efile:///app/vendor/pkg\nrequests==2.31.0\n",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, requirementsNeedLocalFiles(tt.content))
		})
	}
}

func TestInstallNeedsAllFiles(t *testing.T) {
	t.Run("temp fixtures", func(t *testing.T) {
		tests := []struct {
			name         string
			requirements string
			pyproject    string
			want         bool
		}{
			{
				name:         "requirements.txt with relative file wheel",
				requirements: "dummy @ file:_vendor/dummy-1.0.0-py3-none-any.whl\n",
				want:         true,
			},
			{
				name:         "remote requirements.txt ignores pyproject path dep",
				requirements: "requests==2.31.0\n",
				pyproject:    "[tool.poetry.dependencies]\nmy-lib = { path = \"./lib\" }\n",
				want:         false,
			},
			{
				name:      "pyproject tool.uv.sources path dependency",
				pyproject: "[project]\nname = \"app\"\ndependencies = [\"foo\"]\n\n[tool.uv.sources]\nfoo = { path = \"./packages/foo\" }\n",
				want:      true,
			},
			{
				name:      "pyproject local git repository",
				pyproject: "[project]\nname = \"app\"\ndependencies = [\"pkg @ git+file:///app/vendor/pkg\"]\n",
				want:      true,
			},
			{
				name:      "pyproject PEP 621 readme and license file tables",
				pyproject: "[project]\nname = \"app\"\nreadme = {file = \"README.md\"}\nlicense = {file = \"LICENSE\"}\n",
				want:      false,
			},
			{
				name:      "pyproject pythonpath setting",
				pyproject: "[tool.pytest.ini_options]\npythonpath = \"src\"\n",
				want:      false,
			},
			{
				name:      "poetry path dependency",
				pyproject: "[tool.poetry.dependencies]\nlocal-lib = { path = \"./local-lib\" }\n",
				want:      true,
			},
			{
				name:      "uv dotted path dependency without workspace table",
				pyproject: "[project]\nname = \"app\"\nversion = \"0.1.0\"\ndependencies = [\"pkg\"]\n\n[tool.uv.sources]\npkg.path = \"./vendor/pkg\"\n",
				want:      true,
			},
			{
				name:      "poetry dotted path dependency",
				pyproject: "[tool.poetry]\nname = \"app\"\nversion = \"0.1.0\"\n\n[tool.poetry.dependencies]\npython = \">=3.11\"\npkg.path = \"./vendor/pkg\"\n",
				want:      true,
			},
			{
				name:         "remote requirements takes precedence over uv dotted path",
				requirements: "requests==2.31.0\n",
				pyproject:    "[project]\nname = \"app\"\nversion = \"0.1.0\"\ndependencies = [\"pkg\"]\n\n[tool.uv.sources]\npkg.path = \"./vendor/pkg\"\n",
				want:         false,
			},
			{
				name:         "remote requirements takes precedence over poetry dotted path",
				requirements: "requests==2.31.0\n",
				pyproject:    "[tool.poetry]\nname = \"app\"\nversion = \"0.1.0\"\n\n[tool.poetry.dependencies]\npython = \">=3.11\"\npkg.path = \"./vendor/pkg\"\n",
				want:         false,
			},
			{
				name:      "commented dotted path dependency",
				pyproject: "[project]\nname = \"app\"\n# pkg.path = \"./vendor/pkg\"\n",
				want:      false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmpDir := t.TempDir()
				if tt.requirements != "" {
					require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte(tt.requirements), 0o644))
				}
				if tt.pyproject != "" {
					require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(tt.pyproject), 0o644))
				}

				ctx := testingUtils.CreateGenerateContext(t, tmpDir)
				provider := PythonProvider{}
				require.Equal(t, tt.want, provider.installNeedsAllFiles(ctx))
			})
		}
	})

	t.Run("examples", func(t *testing.T) {
		// explicit expectations force each new Python example to make a source-copy decision
		expected := map[string]bool{
			"python-bot-only":              false,
			"python-compiled":              false,
			"python-django":                false,
			"python-fastapi":               false,
			"python-fasthtml":              false,
			"python-flask":                 false,
			"python-freethreaded":          false,
			"python-latest":                false,
			"python-latest-psycopg":        false,
			"python-oldest":                false,
			"python-pdm":                   false,
			"python-pip":                   false,
			"python-pip-vendored-wheel":    true,
			"python-pipfile":               false,
			"python-playwright":            false,
			"python-poetry":                false,
			"python-psycopg-binary":        false,
			"python-system-deps":           false,
			"python-uv":                    false,
			"python-uv-latest-locked":      false,
			"python-uv-packaged":           false,
			"python-uv-tool-versions":      false,
			"python-uv-workspace":          true,
			"python-uv-workspace-postgres": true,
		}

		examplesApp, err := app.NewApp("../../../examples")
		require.NoError(t, err)
		exampleDirs, err := examplesApp.FindDirectories("python-*")
		require.NoError(t, err)
		require.Len(t, expected, len(exampleDirs), "update source-copy expectations when Python examples change")

		for _, name := range exampleDirs {
			t.Run(name, func(t *testing.T) {
				want, ok := expected[name]
				require.True(t, ok, "add a source-copy expectation for %s", name)
				ctx := testingUtils.CreateGenerateContext(t, filepath.Join(examplesApp.Source, name))
				provider := PythonProvider{}
				require.Equal(t, want, provider.installNeedsAllFiles(ctx))

				// assert the planned inputs too, so a correct detector cannot hide unnecessary source copies
				require.NoError(t, provider.Plan(ctx))
				step := ctx.GetStepByName("install")
				require.NotNil(t, step)
				installStep, ok := (*step).(*generate.CommandStepBuilder)
				require.True(t, ok)
				if want {
					require.Contains(t, installStep.Inputs, plan.NewLocalLayer())
				} else {
					require.NotContains(t, installStep.Inputs, plan.NewLocalLayer())
				}
			})
		}
	})
}

func TestInstallStepLocalLayerWiring(t *testing.T) {
	t.Run("local git repository includes local layer in install step", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("pkg @ git+file:///app/vendor/pkg\n"), 0o644))
		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}

		// pip must receive the application source so it can clone the local repository.
		require.NoError(t, provider.Plan(ctx))
		step := ctx.GetStepByName("install")
		require.NotNil(t, step)
		installStep, ok := (*step).(*generate.CommandStepBuilder)
		require.True(t, ok)
		require.Contains(t, installStep.Inputs, plan.NewLocalLayer())
	})

	t.Run("vendored wheel includes local layer in install step", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/python-pip-vendored-wheel")
		provider := PythonProvider{}

		require.NoError(t, provider.Plan(ctx))
		step := ctx.GetStepByName("install")
		require.NotNil(t, step)
		installStep, ok := (*step).(*generate.CommandStepBuilder)
		require.True(t, ok)
		require.Contains(t, installStep.Inputs, plan.NewLocalLayer())
	})

	t.Run("remote requirements only copies requirements.txt in install step", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/python-pip")
		provider := PythonProvider{}

		require.NoError(t, provider.Plan(ctx))
		step := ctx.GetStepByName("install")
		require.NotNil(t, step)
		installStep, ok := (*step).(*generate.CommandStepBuilder)
		require.True(t, ok)
		require.NotContains(t, installStep.Inputs, plan.NewLocalLayer())
		require.Contains(t, installStep.Commands, plan.NewCopyCommand("requirements.txt"))
	})

	t.Run("compact editable requirement includes local layer in install step", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("-efile:///app/vendor/pkg\n"), 0o644))
		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}

		require.NoError(t, provider.Plan(ctx))
		step := ctx.GetStepByName("install")
		require.NotNil(t, step)
		installStep, ok := (*step).(*generate.CommandStepBuilder)
		require.True(t, ok)
		require.Contains(t, installStep.Inputs, plan.NewLocalLayer())
	})

	t.Run("compact find-links requirement includes local layer in install step", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("-ffile:///app/vendor\n"), 0o644))
		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}

		require.NoError(t, provider.Plan(ctx))
		step := ctx.GetStepByName("install")
		require.NotNil(t, step)
		installStep, ok := (*step).(*generate.CommandStepBuilder)
		require.True(t, ok)
		require.Contains(t, installStep.Inputs, plan.NewLocalLayer())
	})

	t.Run("uv dotted path includes local layer in install step", func(t *testing.T) {
		tmpDir := t.TempDir()
		pyproject := "[project]\nname = \"app\"\nversion = \"0.1.0\"\ndependencies = [\"pkg\"]\n\n[tool.uv.sources]\npkg.path = \"./vendor/pkg\"\n"
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(""), 0o644))
		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}

		require.NoError(t, provider.Plan(ctx))
		step := ctx.GetStepByName("install")
		require.NotNil(t, step)
		installStep, ok := (*step).(*generate.CommandStepBuilder)
		require.True(t, ok)
		require.Contains(t, installStep.Inputs, plan.NewLocalLayer())
	})

	t.Run("poetry dotted path includes local layer in install step", func(t *testing.T) {
		tmpDir := t.TempDir()
		pyproject := "[tool.poetry]\nname = \"app\"\nversion = \"0.1.0\"\n\n[tool.poetry.dependencies]\npython = \">=3.11\"\npkg.path = \"./vendor/pkg\"\n"
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "poetry.lock"), []byte(""), 0o644))
		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}

		require.NoError(t, provider.Plan(ctx))
		step := ctx.GetStepByName("install")
		require.NotNil(t, step)
		installStep, ok := (*step).(*generate.CommandStepBuilder)
		require.True(t, ok)
		require.Contains(t, installStep.Inputs, plan.NewLocalLayer())
	})
}

func TestPackageManagerToolProvisioning(t *testing.T) {
	t.Run("requirements + pyproject + uv.lock", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("print('hello')\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("requests==2.31.0\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte("[project]\nname = \"app\"\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(""), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}
		require.NoError(t, provider.Plan(ctx))

		pkgNames := getMisePackageNames(ctx.GetMiseStepBuilder())
		require.Equal(t, []string{"python", "uv"}, pkgNames)
		require.True(t, stepHasExecCommand(ctx, "install", "pip install -r requirements.txt"))
		require.False(t, stepHasExecCommand(ctx, "build", "uv sync --locked --no-dev --no-editable"))
	})

	t.Run("pyproject + uv.lock + poetry.lock", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("print('hello')\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte("[project]\nname = \"app\"\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(""), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "poetry.lock"), []byte(""), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}
		require.NoError(t, provider.Plan(ctx))

		pkgNames := getMisePackageNames(ctx.GetMiseStepBuilder())
		require.Equal(t, []string{"python", "poetry", "uv"}, pkgNames)
		require.True(t, stepHasExecCommand(ctx, "install", "uv sync --locked --no-dev --no-install-project"))
		require.True(t, stepHasExecCommand(ctx, "build", "uv sync --locked --no-dev --no-editable"))
	})

	t.Run("pyproject + pdm.lock + Pipfile", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("print('hello')\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte("[project]\nname = \"app\"\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pdm.lock"), []byte(""), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Pipfile"), []byte("[packages]\n"), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}
		require.NoError(t, provider.Plan(ctx))

		pkgNames := getMisePackageNames(ctx.GetMiseStepBuilder())
		require.Equal(t, []string{"python", "pipx", "uv", "pipx:pdm", "pipx:pipenv"}, pkgNames)
		require.True(t, stepHasExecCommand(ctx, "install", "pdm install --check --prod --no-editable"))
	})

	t.Run("all indicators", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("print('hello')\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("requests==2.31.0\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte("[project]\nname = \"app\"\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(""), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "poetry.lock"), []byte(""), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pdm.lock"), []byte(""), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Pipfile"), []byte("[packages]\n"), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}
		require.NoError(t, provider.Plan(ctx))

		pkgNames := getMisePackageNames(ctx.GetMiseStepBuilder())
		require.Equal(t, []string{"python", "pipx", "uv", "poetry", "pipx:pdm", "pipx:pipenv"}, pkgNames)
		require.True(t, stepHasExecCommand(ctx, "install", "pip install -r requirements.txt"))
	})

	t.Run("main.py + uv.lock without pyproject.toml", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("print('hello')\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(""), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}
		require.NoError(t, provider.Plan(ctx))

		pkgNames := getMisePackageNames(ctx.GetMiseStepBuilder())
		require.Equal(t, []string{"python", "uv"}, pkgNames)

		step := ctx.GetStepByName("install")
		require.NotNil(t, step)
		installStep, ok := (*step).(*generate.CommandStepBuilder)
		require.True(t, ok)
		require.Empty(t, installStep.Commands)
	})

	t.Run("python-only app", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("print('hello')\n"), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}
		require.NoError(t, provider.Plan(ctx))

		pkgNames := getMisePackageNames(ctx.GetMiseStepBuilder())
		require.Equal(t, []string{"python"}, pkgNames)
	})

	t.Run("version override on nonselected detected manager", func(t *testing.T) {
		tmpDir, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("print('hello')\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("requests==2.31.0\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(""), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "mise.toml"), []byte("[tools]\nuv = \"0.4.0\"\n"), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, tmpDir)
		provider := PythonProvider{}
		require.NoError(t, provider.Plan(ctx))

		uvPkg := ctx.GetMiseStepBuilder().Resolver.Get("uv")
		require.NotNil(t, uvPkg)
		require.Equal(t, "0.4.0", uvPkg.Version)
		require.True(t, stepHasExecCommand(ctx, "install", "pip install -r requirements.txt"))
	})
}

func getMisePackageNames(miseStep *generate.MiseStepBuilder) []string {
	names := make([]string, len(miseStep.MisePackages))
	for i, pkg := range miseStep.MisePackages {
		names[i] = pkg.Name
	}
	return names
}
