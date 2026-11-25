package python

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type PackageManager string

const (
	// no supported manager was detected; Python is still installed without a dependency install step
	PackageManagerNone PackageManager = ""

	PackageManagerPip    PackageManager = "pip"
	PackageManagerUv     PackageManager = "uv"
	PackageManagerPoetry PackageManager = "poetry"
	PackageManagerPdm    PackageManager = "pdm"
	PackageManagerPipenv PackageManager = "pipenv"

	UV_CACHE_DIR   = "/opt/uv-cache"
	PIP_CACHE_DIR  = "/opt/pip-cache"
	VENV_PATH      = "/app/.venv"
	LOCAL_BIN_PATH = "/root/.local/bin"
)

func (pm PackageManager) Name() string {
	return string(pm)
}

// requirements.txt takes precedence over pyproject managers, whose lockfiles select the installer
func (p *PythonProvider) detectPackageManager(ctx *generate.GenerateContext) PackageManager {
	if p.hasRequirements(ctx) {
		return PackageManagerPip
	} else if p.hasPyproject(ctx) && p.hasUv(ctx) {
		return PackageManagerUv
	} else if p.hasPyproject(ctx) && p.hasPoetry(ctx) {
		return PackageManagerPoetry
	} else if p.hasPyproject(ctx) && p.hasPdm(ctx) {
		return PackageManagerPdm
	} else if p.hasPipfile(ctx) {
		return PackageManagerPipenv
	}
	return PackageManagerNone
}

// discover all package-manager tools needed by the application's manifests
func (p *PythonProvider) getPackageManagerTools(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) []string {
	var packages []string

	addTool := func(pkg string) {
		if !slices.Contains(packages, pkg) {
			packages = append(packages, pkg)
		}
	}

	if p.hasPdm(ctx) || p.hasPipfile(ctx) {
		miseStep.Default("pipx", "latest")
		addTool("pipx")

		miseStep.Default("uv", "latest")
		addTool("uv")
	}

	if p.hasPoetry(ctx) {
		miseStep.Default("poetry", "latest")
		addTool("poetry")
	}

	if p.hasPdm(ctx) {
		// Keep the pipx: alias so existing mise.toml version overrides still match this tool.
		// TODO(breaking): Migrate PDM and Pipenv to pypi: and update version override handling.
		miseStep.Default("pipx:pdm", "latest")
		addTool("pipx:pdm")
	}

	if p.hasUv(ctx) {
		miseStep.Default("uv", "latest")
		addTool("uv")
	}

	if p.hasPipfile(ctx) {
		// Like pypi:, this backend uses uv when available and falls back to pipx.
		miseStep.Default("pipx:pipenv", "latest")
		addTool("pipx:pipenv")
	}

	return packages
}

// return the environment paths that must be copied into the deploy image to retain installed dependencies
func (pm PackageManager) Install(p *PythonProvider, ctx *generate.GenerateContext, install *generate.CommandStepBuilder) []string {
	switch pm {
	case PackageManagerPip:
		return p.InstallPip(ctx, install)
	case PackageManagerUv:
		return p.InstallUv(ctx, install)
	case PackageManagerPoetry:
		return p.InstallPoetry(ctx, install)
	case PackageManagerPdm:
		return p.InstallPDM(ctx, install)
	case PackageManagerPipenv:
		return p.InstallPipenv(ctx, install)
	default:
		// Plain Python scripts can run without a detected package manager or installed dependencies.
		return []string{}
	}
}

func (pm PackageManager) ConfigureBuildStep(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	if pm == PackageManagerUv {
		// the project is not installed during the install phase, because it requires the project source
		build.AddCommands([]plan.Command{
			plan.NewExecCommand("uv sync --locked --no-dev --no-editable"),
		})
	}
}

func (p *PythonProvider) InstallUv(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) []string {
	ctx.Logger.LogInfo("Using uv")

	install.AddCache(ctx.Caches.AddCache("uv", UV_CACHE_DIR))
	install.AddEnvVars(map[string]string{
		"UV_COMPILE_BYTECODE": "1",
		"UV_LINK_MODE":        "copy",
		"UV_CACHE_DIR":        UV_CACHE_DIR,
		"UV_PYTHON_DOWNLOADS": "never",
		"VIRTUAL_ENV":         VENV_PATH,
	})

	install.AddEnvVars(p.GetPythonEnvVars(ctx))

	p.copyInstallFiles(ctx, install)
	installCommands := []plan.Command{
		plan.NewPathCommand(LOCAL_BIN_PATH),
		plan.NewPathCommand(VENV_PATH + "/bin"),
		// if we exclude workspace packages, uv.lock will fail the frozen test and the user will get an error
		// to avoid this, we (a) detect if workspace packages are required (b) if they aren't, we don't include project
		// source in order to optimize layer caching (c) install project in the build phase.
		plan.NewExecCommand("uv sync --locked --no-dev --no-install-project"),
	}

	install.AddCommands(installCommands)

	return []string{VENV_PATH}
}

func (p *PythonProvider) InstallPipenv(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) []string {
	ctx.Logger.LogInfo("Using pipenv")

	install.AddEnvVars(p.GetPythonEnvVars(ctx))
	install.AddEnvVars(map[string]string{
		"PIPENV_CHECK_UPDATE":       "false",
		"PIPENV_VENV_IN_PROJECT":    "1",
		"PIPENV_IGNORE_VIRTUALENVS": "1",
	})

	install.AddCommands([]plan.Command{
		plan.NewPathCommand(LOCAL_BIN_PATH),
		plan.NewPathCommand(VENV_PATH + "/bin"),
	})

	if ctx.App.HasFile("Pipfile.lock") {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("Pipfile"),
			plan.NewCopyCommand("Pipfile.lock"),
			plan.NewExecCommand("pipenv install --deploy --ignore-pipfile"),
		})
	} else {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("Pipfile"),
			plan.NewExecCommand("pipenv install --skip-lock"),
		})
	}

	return []string{VENV_PATH}
}

func (p *PythonProvider) InstallPDM(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) []string {
	ctx.Logger.LogInfo("Using pdm")

	install.AddEnvVars(p.GetPythonEnvVars(ctx))
	install.AddEnvVars(map[string]string{
		"PDM_CHECK_UPDATE": "false",
	})

	p.copyInstallFiles(ctx, install)
	installCommands := []plan.Command{
		plan.NewPathCommand(LOCAL_BIN_PATH),
		plan.NewPathCommand(VENV_PATH + "/bin"),
		plan.NewExecCommand("pdm install --check --prod --no-editable"),
	}

	install.AddCommands(installCommands)

	return []string{VENV_PATH}
}

// poetry must target the same environment that is copied into the deploy image
func (p *PythonProvider) InstallPoetry(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) []string {
	ctx.Logger.LogInfo("Using poetry")

	install.AddEnvVars(p.GetPythonEnvVars(ctx))
	install.AddEnvVars(map[string]string{
		"VIRTUAL_ENV":                   VENV_PATH,
		"POETRY_VIRTUALENVS_PATH":       VENV_PATH,
		"POETRY_VIRTUALENVS_IN_PROJECT": "true",
	})

	p.copyInstallFiles(ctx, install)
	installCommands := []plan.Command{
		plan.NewPathCommand(LOCAL_BIN_PATH),
		plan.NewPathCommand(VENV_PATH + "/bin"),
		plan.NewExecCommand("poetry install --no-interaction --no-ansi --only main --no-root"),
	}

	install.AddCommands(installCommands)

	return []string{VENV_PATH}
}

// pip needs an explicit virtual environment so dependencies can be copied into the deploy image
func (p *PythonProvider) InstallPip(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) []string {
	ctx.Logger.LogInfo("Using pip")

	install.AddCache(ctx.Caches.AddCache("pip", PIP_CACHE_DIR))
	install.AddEnvVars(p.GetPythonEnvVars(ctx))
	install.AddEnvVars(map[string]string{
		"PIP_CACHE_DIR": PIP_CACHE_DIR,
		"VIRTUAL_ENV":   VENV_PATH,
	})

	install.AddCommands([]plan.Command{
		plan.NewExecCommand(fmt.Sprintf("python -m venv %s", VENV_PATH)),
		plan.NewPathCommand(VENV_PATH + "/bin"),
	})
	p.copyInstallFiles(ctx, install)
	install.AddCommands([]plan.Command{
		plan.NewExecCommand("pip install -r requirements.txt"),
	})

	return []string{VENV_PATH}
}

// keep installs cached across source changes unless local dependencies require application files
func (p *PythonProvider) copyInstallFiles(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) {
	if p.installNeedsAllFiles(ctx) {
		install.AddInput(plan.NewLocalLayer())
		return
	}

	files := []string{
		"requirements.txt",
		"pyproject.toml",
		"Pipfile",
		"poetry.lock",
		"uv.lock",
		"pdm.lock",
	}

	for _, file := range files {
		if ctx.App.HasFile(file) {
			install.AddCommand(plan.NewCopyCommand(file))
		}
	}
}

// Detect local dependencies in requirements.txt and pyproject.toml so installs receive application source.
var (
	// Match local file URLs, e.g. `pkg @ file:_vendor/pkg.whl` or `git+file:///app/lib`.
	localFileSchemeRegex = regexp.MustCompile(`(?:^|[\s@='"])(?:(?:git|svn|hg|bzr)\+)?file:`)

	// Editable requirements such as `-e ./lib` need source available during installation.
	editableRequirementRegex = regexp.MustCompile(`^(?:-e(?:\s*|=)?|--editable(?:\s+|=))(\S*)`)

	// Capture the wheel source in `-f _vendor` or `--find-links=https://example.com/wheels`.
	findLinksOptionRegex = regexp.MustCompile(`^(?:-f(?:\s*|=)?|--find-links(?:\s+|=))(\S+)`)

	// Match direct local references such as `pkg @ ./lib` and `pkg @ /app/lib`.
	pep508LocalPathRegex = regexp.MustCompile(`@\s*(?:\.\.?/|/)`)

	// Recognize wheel and source archive paths such as `_vendor/pkg.tar.gz`.
	localArchiveExtensionRegex = regexp.MustCompile(`\.(?:whl|tar\.gz|tgz|zip)(?:$|\s|;|#)`)

	// Exclude downloadable archives and find-links URLs from local-file detection.
	remoteRequirementSchemeRegex = regexp.MustCompile(`(?i)(?:https?|ssh)://|(?:git|svn|hg|bzr)\+`)

	// Match `path = "./lib"` and `pkg.path = "./lib"`, excluding `pythonpath = "src"`.
	pyprojectLocalPathKeyRegex = regexp.MustCompile(`(?:^|[{\s,.])path\s*=`)
)

// Inspect python dependency files and determine if local packages are referenced, and therefore all files are required
// for installation.
func (p *PythonProvider) installNeedsAllFiles(ctx *generate.GenerateContext) bool {
	if requirementsContent, err := ctx.App.ReadFile("requirements.txt"); err == nil {
		return requirementsNeedLocalFiles(requirementsContent)
	}

	// inspect pyproject.toml for local path references or uv workspace usage
	if pyprojectContent, err := ctx.App.ReadFile("pyproject.toml"); err == nil {
		if pyprojectNeedsLocalFiles(pyprojectContent) {
			return true
		}
	}

	// Conservatively copy workspace sources even when no member is an explicit dependency.
	// A pointer distinguishes an absent tool.uv.workspace key from an empty workspace table.
	var pyproject struct {
		Tool struct {
			Uv struct {
				Workspace *struct{} `toml:"workspace"`
			} `toml:"uv"`
		} `toml:"tool"`
	}
	if err := ctx.App.ReadTOML("pyproject.toml", &pyproject); err == nil {
		return pyproject.Tool.Uv.Workspace != nil
	} else {
		log.Infof("Failed to read pyproject.toml: %v", err)
	}

	return false
}

// local packages and find-links directories need application files available during pip installation
func requirementsNeedLocalFiles(content string) bool {
	for rawLine := range strings.Lines(content) {
		line := strings.TrimSpace(rawLine)
		// Comments may mention local paths without adding a dependency.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// File URLs, editable installs, and direct paths require the application's source.
		if localFileSchemeRegex.MatchString(line) ||
			editableRequirementRegex.MatchString(line) ||
			pep508LocalPathRegex.MatchString(line) {
			return true
		}

		// A local wheel directory must be copied; a remote find-links index needs no source.
		if matches := findLinksOptionRegex.FindStringSubmatch(line); len(matches) > 1 {
			if !remoteRequirementSchemeRegex.MatchString(matches[1]) {
				return true
			}
		}

		// Bare archive paths need source, whereas remote archive URLs can be downloaded.
		if localArchiveExtensionRegex.MatchString(line) && !remoteRequirementSchemeRegex.MatchString(line) {
			return true
		}
	}

	return false
}

// dependency paths require application source, while readme and license file tables alone do not
func pyprojectNeedsLocalFiles(content string) bool {
	for rawLine := range strings.Lines(content) {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if localFileSchemeRegex.MatchString(line) || pyprojectLocalPathKeyRegex.MatchString(line) {
			return true
		}
	}

	return false
}

func (p *PythonProvider) hasRequirements(ctx *generate.GenerateContext) bool {
	return ctx.App.HasFile("requirements.txt")
}

func (p *PythonProvider) hasPyproject(ctx *generate.GenerateContext) bool {
	return ctx.App.HasFile("pyproject.toml")
}

func (p *PythonProvider) hasPipfile(ctx *generate.GenerateContext) bool {
	return ctx.App.HasFile("Pipfile")
}

func (p *PythonProvider) hasPoetry(ctx *generate.GenerateContext) bool {
	return ctx.App.HasFile("poetry.lock")
}

func (p *PythonProvider) hasPdm(ctx *generate.GenerateContext) bool {
	return ctx.App.HasFile("pdm.lock")
}

func (p *PythonProvider) hasUv(ctx *generate.GenerateContext) bool {
	return ctx.App.HasFile("uv.lock")
}
