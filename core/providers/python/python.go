package python

import (
	"fmt"
	"maps"
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/internal/utils"
)

const (
	DEFAULT_PYTHON_VERSION = "3.13"
	PLAYWRIGHT_CACHE_DIR   = "/root/.cache/ms-playwright"
	PLAYWRIGHT_INSTALL_VAR = "PYTHON_PLAYWRIGHT_INSTALL"
)

// Keep this aligned with Playwright's Chromium deps in nativeDeps.ts:
// https://github.com/microsoft/playwright/blob/main/packages/playwright-core/src/server/registry/nativeDeps.ts
var pythonPlaywrightRuntimeDependencies = []string{
	"libasound2",
	"libatk-bridge2.0-0",
	"libatk1.0-0",
	"libatspi2.0-0",
	"libcairo2",
	"libcups2",
	"libdbus-1-3",
	"libdrm2",
	"libgbm1",
	"libglib2.0-0",
	"libnspr4",
	"libnss3",
	"libpango-1.0-0",
	"libx11-6",
	"libxcb1",
	"libxcomposite1",
	"libxdamage1",
	"libxext6",
	"libxfixes3",
	"libxkbcommon0",
	"libxrandr2",
}

var pythonDefaultExcludePatterns = []string{".venv", "venv"}

type PythonProvider struct{}

func (p *PythonProvider) Name() string {
	return "python"
}

func (p *PythonProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *PythonProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	hasPython := p.getMainPythonFile(ctx) != "" ||
		p.hasRequirements(ctx) ||
		p.hasPyproject(ctx) ||
		p.hasPipfile(ctx)

	return hasPython, nil
}

func (p *PythonProvider) Plan(ctx *generate.GenerateContext) error {
	ctx.SetProviderDefaultExcludes(pythonDefaultExcludePatterns)

	pm := p.detectPackageManager(ctx)
	p.InstallMisePackages(ctx, ctx.GetMiseStepBuilder())

	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepLayer(p.GetBuilderDeps(ctx).Name()))

	install.Secrets = []string{}
	install.UseSecretsWithPrefixes([]string{"PYTHON", "PIP", "PIPX", "UV", "PDM", "POETRY"})

	build := ctx.NewCommandStep("build")
	installOutputs := pm.Install(p, ctx, install)
	pm.ConfigureBuildStep(ctx, build)

	usesPlaywright := p.usesProductionDep(ctx, "playwright")
	installPlaywright := ctx.Env.IsConfigVariableTruthy(PLAYWRIGHT_INSTALL_VAR)

	// Automatic browser installation caused issues for an existing user, so keep this opt-in.
	if usesPlaywright && !installPlaywright {
		ctx.Logger.LogSuggestion(
			"Set `RAILPACK_PYTHON_PLAYWRIGHT_INSTALL=1` to install Playwright browsers",
			"/languages/python#playwright",
		)
	}

	if installPlaywright {
		ctx.Logger.LogInfo("Installing Playwright chromium browser")
		// --only-shell installs the smaller Chromium headless shell, which is
		// more appropriate for server environments than full Chromium.
		install.AddCommand(plan.NewExecCommand("playwright install --only-shell"))
		// Include the browser cache so its binaries are available in the deploy stage.
		installOutputs = append(installOutputs, PLAYWRIGHT_CACHE_DIR)
	}

	p.addMetadata(ctx, pm)

	build.AddInput(plan.NewStepLayer(install.Name()))
	build.AddInput(plan.NewLocalLayer())

	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)
	maps.Copy(ctx.Deploy.Variables, p.GetPythonEnvVars(ctx))

	installArtifacts := plan.NewStepLayer(build.Name(), plan.Filter{
		Include: installOutputs,
	})

	p.AddRuntimeDeps(ctx)

	ctx.Deploy.AddInputs([]plan.Layer{
		ctx.GetMiseStepBuilder().GetLayer(),
		installArtifacts,
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{"."},
			Exclude: []string{strings.TrimPrefix(VENV_PATH, "/app/")},
		}),
	})

	return nil
}

func (p *PythonProvider) GetStartCommand(ctx *generate.GenerateContext) string {
	startCommand := ""

	if p.isDjango(ctx) {
		startCommand = p.getDjangoStartCommand(ctx)
	}

	mainPythonFile := p.getMainPythonFile(ctx)
	hasMainPythonFile := mainPythonFile != ""

	if p.isFasthtml(ctx) && hasMainPythonFile && p.usesDep(ctx, "uvicorn") {
		startCommand = "uvicorn main:app --host 0.0.0.0 --port ${PORT:-8000}"
	}

	if p.isFastapi(ctx) && hasMainPythonFile && p.usesDep(ctx, "uvicorn") {
		startCommand = "uvicorn main:app --host 0.0.0.0 --port ${PORT:-8000}"
	}

	if p.isFlask(ctx) && hasMainPythonFile && p.usesDep(ctx, "gunicorn") {
		startCommand = "gunicorn --bind 0.0.0.0:${PORT:-8000} main:app"
	}

	if startCommand == "" && hasMainPythonFile {
		startCommand = fmt.Sprintf("python %s", mainPythonFile)
	}

	return startCommand
}

func (p *PythonProvider) getMainPythonFile(ctx *generate.GenerateContext) string {
	for _, file := range []string{"main.py", "app.py", "start.py", "bot.py", "hello.py", "server.py"} {
		if ctx.App.HasFile(file) {
			return file
		}
	}
	return ""
}

func (p *PythonProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *PythonProvider) StartCommandHelp() string {
	return "To start your Python application, Railpack will automatically:\n\n" +
		"1. Start FastHTML projects with uvicorn\n" +
		"2. Start FastAPI projects with uvicorn\n" +
		"3. Start Flask projects with gunicorn\n" +
		"4. Start Django projects with the gunicorn production server\n\n" +
		"Otherwise, it will run the main.py or app.py file in your project root"
}

func (p *PythonProvider) AddRuntimeDeps(ctx *generate.GenerateContext) {
	for dep, requiredPkgs := range pythonRuntimeDepRequirements {
		if p.usesDep(ctx, dep) {
			ctx.Logger.LogInfo("Installing runtime apt packages for %s: %v", dep, requiredPkgs)
			ctx.Deploy.AddAptPackages(requiredPkgs)
		}
	}

	if ctx.Env.IsConfigVariableTruthy(PLAYWRIGHT_INSTALL_VAR) {
		ctx.Logger.LogInfo("Installing runtime apt packages for playwright: %v", pythonPlaywrightRuntimeDependencies)
		ctx.Deploy.AddAptPackages(pythonPlaywrightRuntimeDependencies)
	}

	if p.usesPostgres(ctx) {
		ctx.Deploy.AddAptPackages([]string{"libpq5"})
	}

	if p.usesMysql(ctx) {
		ctx.Deploy.AddAptPackages([]string{"default-mysql-client"})
	}
}

func (p *PythonProvider) GetBuilderDeps(ctx *generate.GenerateContext) *generate.MiseStepBuilder {
	miseStep := ctx.GetMiseStepBuilder()

	// certain packages require apt libraries in order to properly build. We shouldn't handle all cases, but we attempt
	// to cover as many popular packages as possible.
	for dep, requiredPkgs := range pythonBuildDepRequirements {
		if p.usesDep(ctx, dep) {
			ctx.Logger.LogInfo("Installing build apt packages for %s: %v", dep, requiredPkgs)
			miseStep.SupportingAptPackages = append(miseStep.SupportingAptPackages, requiredPkgs...)
		}
	}

	// detecting database support is multi-faceted, so we special case them
	// note that these packages do *not* persist past the build phase and must be re-installed in the runtime if needed
	if p.usesPostgres(ctx) {
		miseStep.SupportingAptPackages = append(miseStep.SupportingAptPackages, "libpq-dev")
	}

	if p.usesMysql(ctx) {
		miseStep.SupportingAptPackages = append(miseStep.SupportingAptPackages, "default-libmysqlclient-dev")
	}

	return miseStep
}

func (p *PythonProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	python := miseStep.Default("python", DEFAULT_PYTHON_VERSION)

	// NOTE: Version resolution precedence matters here.
	// We evaluate manifest files (Pipfile, runtime.txt) first to establish the baseline.
	if pipfileVersion, pipfileVarName := parseVersionFromPipfile(ctx); pipfileVersion != "" {
		miseStep.Version(python, pipfileVersion, fmt.Sprintf("Pipfile > %s", pipfileVarName))
	}

	if runtimeFile, err := ctx.App.ReadFile("runtime.txt"); err == nil {
		miseStep.Version(python, utils.ExtractSemverVersion(string(runtimeFile)), "runtime.txt")
	}

	// Collect all packages that will be used by the provider
	packages := []string{"python"}

	packages = append(packages, p.getPackageManagerTools(ctx, miseStep)...)

	miseStep.UseMiseVersions(ctx, packages)

	if envVersion, varName := ctx.Env.GetConfigVariable("PYTHON_VERSION"); envVersion != "" {
		miseStep.Version(python, envVersion, varName)
	}

	// Disable Python compilation to avoid incompatibility issues with some packages
	// https://mise.jdx.dev/lang/python.html#python.compile
	miseStep.AddMiseSetting("python.compile", false)
}

func (p *PythonProvider) GetPythonEnvVars(ctx *generate.GenerateContext) map[string]string {
	return map[string]string{
		"PYTHONFAULTHANDLER":            "1",
		"PYTHONUNBUFFERED":              "1",
		"PYTHONHASHSEED":                "random",
		"PYTHONDONTWRITEBYTECODE":       "1",
		"PIP_DISABLE_PIP_VERSION_CHECK": "1",
		"PIP_DEFAULT_TIMEOUT":           "100",
	}
}

func (p *PythonProvider) usesBinaryPsycopg(ctx *generate.GenerateContext) bool {
	return p.usesDep(ctx, "psycopg2-binary") || p.usesDep(ctx, "psycopg[binary]")
}

func (p *PythonProvider) usesPostgres(ctx *generate.GenerateContext) bool {
	if p.usesBinaryPsycopg(ctx) {
		return false
	}

	djangoPythonRe := regexp.MustCompile(`django.db.backends.postgresql`)
	containsDjangoPostgres := len(ctx.App.FindFilesWithContent("**/*.py", djangoPythonRe)) > 0
	return p.usesDep(ctx, "psycopg2") || p.usesDep(ctx, "psycopg") || containsDjangoPostgres
}

func (p *PythonProvider) usesMysql(ctx *generate.GenerateContext) bool {
	djangoPythonRe := regexp.MustCompile(`django.db.backends.mysql`)
	containsDjangoMysql := len(ctx.App.FindFilesWithContent("**/*.py", djangoPythonRe)) > 0
	return p.usesDep(ctx, "mysqlclient") || containsDjangoMysql
}

func (p *PythonProvider) addMetadata(ctx *generate.GenerateContext, pm PackageManager) {
	pkgManager := pm.Name()
	if pkgManager == "" || pkgManager == "pipenv" {
		pkgManager = "pip"
	}

	ctx.Metadata.Set("pythonPackageManager", pkgManager)
	ctx.Metadata.Set("pythonRuntime", p.getRuntime(ctx))
}

// TODO this is incredibly naive: we should parse the files we can distinguish between prod and dev
func (p *PythonProvider) usesDep(ctx *generate.GenerateContext, dep string) bool {
	files, err := ctx.App.FindFiles("**/{requirements.txt,pyproject.toml,Pipfile}")
	if err != nil {
		return false
	}
	for _, file := range files {
		if contents, err := ctx.App.ReadFile(file); err == nil {
			if strings.Contains(strings.ToLower(contents), strings.ToLower(dep)) {
				return true
			}
		}
	}
	return false
}

var pipfileFullVersionRegex = regexp.MustCompile(`python_full_version\s*=\s*['"]([0-9.]*)"?`)
var pipfileShortVersionRegex = regexp.MustCompile(`python_version\s*=\s*['"]([0-9.]*)"?`)

func parseVersionFromPipfile(ctx *generate.GenerateContext) (string, string) {
	pipfile, err := ctx.App.ReadFile("Pipfile")
	if err != nil {
		return "", ""
	}

	if matches := pipfileFullVersionRegex.FindStringSubmatch(string(pipfile)); len(matches) > 1 {
		return matches[1], "python_full_version"
	}

	if matches := pipfileShortVersionRegex.FindStringSubmatch(string(pipfile)); len(matches) > 1 {
		return matches[1], "python_version"
	}

	return "", ""
}

func (p *PythonProvider) isFasthtml(ctx *generate.GenerateContext) bool {
	return p.usesDep(ctx, "python-fasthtml")
}

func (p *PythonProvider) isFlask(ctx *generate.GenerateContext) bool {
	return p.usesDep(ctx, "flask")
}

func (p *PythonProvider) isFastapi(ctx *generate.GenerateContext) bool {
	return p.usesDep(ctx, "fastapi")
}

func (p *PythonProvider) getRuntime(ctx *generate.GenerateContext) string {
	if p.isDjango(ctx) {
		return "django"
	} else if p.isFlask(ctx) {
		return "flask"
	} else if p.isFastapi(ctx) {
		return "fastapi"
	} else if p.isFasthtml(ctx) {
		return "fasthtml"
	}

	return "python"
}

// Mapping of python dependencies to required apt packages

var pythonBuildDepRequirements = map[string][]string{
	"pycairo": {"libcairo2-dev"},
}

var pythonRuntimeDepRequirements = map[string][]string{
	"pycairo":   {"libcairo2"},
	"pdf2image": {"poppler-utils"},
	"pydub":     {"ffmpeg"},
	"pymovie":   {"ffmpeg", "qt5-qmake", "qtbase5-dev", "qtbase5-dev-tools", "qttools5-dev-tools", "libqt5core5a", "python3-pyqt5"},
}
