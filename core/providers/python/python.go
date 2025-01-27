package python

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/railwayapp/railpack-go/core/generate"
	"github.com/railwayapp/railpack-go/core/plan"
)

const (
	DEFAULT_PYTHON_VERSION = "3.13"
	UV_CACHE_DIR           = "/opt/uv-cache"
)

type PythonProvider struct{}

func (p *PythonProvider) Name() string {
	return "python"
}

func (p *PythonProvider) Plan(ctx *generate.GenerateContext) (bool, error) {
	hasPython := ctx.App.HasMatch("main.py") ||
		p.hasRequirements(ctx) ||
		p.hasPyproject(ctx) ||
		p.hasPoetry(ctx) ||
		p.hasPdm(ctx)

	if !hasPython {
		return false, nil
	}

	if err := p.packages(ctx); err != nil {
		return false, err
	}

	if err := p.install(ctx); err != nil {
		return false, err
	}

	if err := p.start(ctx); err != nil {
		return false, err
	}

	p.addMetadata(ctx)

	return false, nil
}

func (p *PythonProvider) start(ctx *generate.GenerateContext) error {
	ctx.Start.Paths = append(ctx.Start.Paths, ".")

	var startCommand string

	if ctx.App.HasMatch("main.py") {
		startCommand = "python main.py"
	}

	if startCommand != "" {
		ctx.Start.Command = startCommand
	}

	return nil
}

// Mapping of python dependencies to required apt packages
var pythonDepRequirements = map[string][]string{
	"cairo":     {"libcairo2-dev"},
	"pdf2image": {"poppler-utils"},
	"pydub":     {"ffmpeg"},
	"pymovie":   {"ffmpeg", "qt5-qmake", "qtbase5-dev", "qtbase5-dev-tools", "qttools5-dev-tools", "libqt5core5a", "python3-pyqt5"},
}

func (p *PythonProvider) install(ctx *generate.GenerateContext) error {
	install := ctx.NewCommandStep("install")
	install.AddCommands([]plan.Command{
		plan.NewPathCommand("/root/.local/bin"),
	})

	hasRequirements := p.hasRequirements(ctx)
	hasPyproject := p.hasPyproject(ctx)
	hasPipfile := p.hasPipfile(ctx)
	hasPoetry := p.hasPoetry(ctx)
	hasPdm := p.hasPdm(ctx)
	hasUv := p.hasUv(ctx)

	install.AddCommands([]plan.Command{
		plan.NewVariableCommand("PYTHONFAULTHANDLER", "1"),
		plan.NewVariableCommand("PYTHONUNBUFFERED", "1"),
		plan.NewVariableCommand("PYTHONHASHSEED", "random"),
		plan.NewVariableCommand("PYTHONDONTWRITEBYTECODE", "1"),
		plan.NewVariableCommand("PIP_NO_CACHE_DIR", "1"),
		plan.NewVariableCommand("PIP_DISABLE_PIP_VERSION_CHECK", "1"),
		plan.NewVariableCommand("PIP_DEFAULT_TIMEOUT", "100"),
		plan.NewExecCommand("pip install --upgrade setuptools"),
	})

	if hasRequirements {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("requirements.txt"),
			plan.NewExecCommand("pip install -r requirements.txt"),
		})
	} else if hasPyproject && hasPoetry {
		install.AddCommands([]plan.Command{
			plan.NewExecCommand("pipx install poetry"),
			plan.NewExecCommand("poetry config virtualenvs.create false"),
			plan.NewCopyCommand("pyproject.toml"),
			plan.NewCopyCommand("poetry.lock"),
			plan.NewExecCommand("poetry install --no-interaction --no-ansi --no-root"),
		})
	} else if hasPyproject && hasPdm {
		// TODO: Fix this. PDM is not working because the packages are installed into a venv
		// that is not available to python at runtime
		install.AddCommands([]plan.Command{
			plan.NewExecCommand("pipx install pdm"),
			plan.NewVariableCommand("PDM_CHECK_UPDATE", "false"),
			plan.NewCopyCommand("pyproject.toml"),
			plan.NewCopyCommand("pdm.lock"),
			plan.NewCopyCommand("."),
			plan.NewExecCommand("pdm install --check --prod --no-editable"),
			plan.NewPathCommand("/app/.venv/bin"),
		})
	} else if hasPyproject && hasUv {
		install.AddCommands([]plan.Command{
			plan.NewVariableCommand("UV_COMPILE_BYTECODE", "1"),
			plan.NewVariableCommand("UV_LINK_MODE", "copy"),
			plan.NewVariableCommand("UV_CACHE_DIR", UV_CACHE_DIR),
			plan.NewExecCommand("pipx install uv"),
			plan.NewCopyCommand("pyproject.toml"),
			plan.NewCopyCommand("uv.lock"),
			plan.NewExecCommand("uv sync --frozen --no-install-project --no-install-workspace --no-dev"),
			plan.NewCopyCommand("."),
			plan.NewExecCommand("uv sync --frozen --no-dev"),
			plan.NewPathCommand("/app/.venv/bin"),
		})
	} else if hasPipfile {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("Pipfile"),
		})

		if ctx.App.HasMatch("Pipfile.lock") {
			install.AddCommands([]plan.Command{
				plan.NewCopyCommand("Pipfile.lock"),
				plan.NewExecCommand("pipenv install --deploy"),
			})
		} else {
			install.AddCommands([]plan.Command{
				plan.NewExecCommand("pipenv install --skip-lock"),
			})
		}
	}

	aptStep := ctx.NewAptStepBuilder("packages:apt")
	aptStep.Packages = []string{"python3-distutils", "gcc", "pkg-config"}
	install.DependsOn = append(install.DependsOn, aptStep.DisplayName)

	for dep, requiredPkgs := range pythonDepRequirements {
		if p.usesDep(ctx, dep) {
			aptStep.Packages = append(aptStep.Packages, requiredPkgs...)
		}
	}

	return nil
}

func (p *PythonProvider) packages(ctx *generate.GenerateContext) error {
	packages := ctx.GetMiseStepBuilder()

	python := packages.Default("python", DEFAULT_PYTHON_VERSION)

	if envVersion := ctx.Env.GetConfigVariable("PYTHON_VERSION"); envVersion != "" {
		packages.Version(python, envVersion, "RAILPACK_PYTHON_VERSION")
	}

	if versionFile, err := ctx.App.ReadFile(".python-version"); err == nil {
		packages.Version(python, string(versionFile), ".python-version")
	}

	if runtimeFile, err := ctx.App.ReadFile("runtime.txt"); err == nil {
		packages.Version(python, string(runtimeFile), "runtime.txt")
	}

	if pipfileVersion := parseVersionFromPipfile(ctx); pipfileVersion != "" {
		packages.Version(python, pipfileVersion, "Pipfile")
	}

	if p.hasPoetry(ctx) || p.hasUv(ctx) || p.hasPdm(ctx) {
		packages.Default("pipx", "latest")
	}

	return nil
}

func (p *PythonProvider) addMetadata(ctx *generate.GenerateContext) {
	hasPoetry := p.hasPoetry(ctx)
	hasPdm := p.hasPdm(ctx)
	hasUv := p.hasUv(ctx)

	pkgManager := "pip"

	if hasPoetry {
		pkgManager = "poetry"
	} else if hasPdm {
		pkgManager = "pdm"
	} else if hasUv {
		pkgManager = "uv"
	}

	ctx.Metadata.Set("packageManager", pkgManager)
	ctx.Metadata.Set("hasRequirements", strconv.FormatBool(p.hasRequirements(ctx)))
	ctx.Metadata.Set("hasPyproject", strconv.FormatBool(p.hasPyproject(ctx)))
	ctx.Metadata.Set("hasPipfile", strconv.FormatBool(p.hasPipfile(ctx)))
}

func (p *PythonProvider) usesDep(ctx *generate.GenerateContext, dep string) bool {
	for _, file := range []string{"requirements.txt", "pyproject.toml", "Pipfile"} {
		if contents, err := ctx.App.ReadFile(file); err == nil {
			if strings.Contains(strings.ToLower(contents), strings.ToLower(dep)) {
				return true
			}
		}
	}
	return false
}

var pipfileVersionRegex = regexp.MustCompile(`(python_version|python_full_version)\s*=\s*['"]([0-9.]*)"?`)

func parseVersionFromPipfile(ctx *generate.GenerateContext) string {
	pipfile, err := ctx.App.ReadFile("Pipfile")
	if err != nil {
		return ""
	}

	matches := pipfileVersionRegex.FindStringSubmatch(string(pipfile))

	if len(matches) > 2 {
		return matches[2]
	}
	return ""
}

func (p *PythonProvider) hasRequirements(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("requirements.txt")
}

func (p *PythonProvider) hasPyproject(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("pyproject.toml")
}

func (p *PythonProvider) hasPipfile(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("Pipfile")
}

func (p *PythonProvider) hasPoetry(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("poetry.lock")
}

func (p *PythonProvider) hasPdm(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("pdm.lock")
}

func (p *PythonProvider) hasUv(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("uv.lock")
}
