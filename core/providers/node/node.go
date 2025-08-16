package node

import (
	"fmt"
	"maps"
	"path"
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type PackageManager string

const (
	DEFAULT_NODE_VERSION = "22"
	DEFAULT_BUN_VERSION  = "latest"

	COREPACK_HOME = "/opt/corepack"
)

var (
	// bunCommandRegex matches "bun" or "bunx" as a command (not part of another word)
	bunCommandRegex = regexp.MustCompile(`(^|\s|;|&|&&|\||\|\|)bunx?\s`)
)

type NodeProvider struct {
	packageJson    *PackageJson
	packageManager PackageManager
	workspace      *Workspace
}

func (p *NodeProvider) Name() string {
	return "node"
}

func (p *NodeProvider) Initialize(ctx *generate.GenerateContext) error {
	packageJson, err := p.GetPackageJson(ctx.App)
	if err != nil {
		return err
	}
	p.packageJson = packageJson

	p.packageManager = p.getPackageManager(ctx.App)

	workspace, err := NewWorkspace(ctx.App)
	if err != nil {
		return err
	}
	p.workspace = workspace

	return nil
}

func (p *NodeProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasMatch("package.json"), nil
}

func (p *NodeProvider) Plan(ctx *generate.GenerateContext) error {
	if p.packageJson == nil {
		return fmt.Errorf("package.json not found")
	}

	p.SetNodeMetadata(ctx)

	ctx.Logger.LogInfo("Using %s package manager", p.packageManager)

	if p.workspace != nil && len(p.workspace.Packages) > 0 {
		ctx.Logger.LogInfo("Found workspace with %d packages", len(p.workspace.Packages))
	}

	isSPA := p.isSPA(ctx)

	miseStep := ctx.GetMiseStepBuilder()
	p.InstallMisePackages(ctx, miseStep)

	// Install
	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepLayer(miseStep.Name()))
	p.InstallNodeDeps(ctx, install)

	// Prune
	prune := ctx.NewCommandStep("prune")
	prune.AddInput(plan.NewStepLayer(install.Name()))
	prune.Secrets = []string{}
	if p.shouldPrune(ctx) && !isSPA {
		p.PruneNodeDeps(ctx, prune)
	}

	// Build
	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(install.Name()))
	p.Build(ctx, build)

	// Deploy
	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)
	maps.Copy(ctx.Deploy.Variables, p.GetNodeEnvVars(ctx))

	// Custom deploy for SPA's
	if isSPA {
		err := p.DeploySPA(ctx, build)
		return err
	}

	// All the files we need to include in the deploy
	buildIncludeDirs := []string{"/root/.cache", "."}

	if p.usesCorepack() {
		buildIncludeDirs = append(buildIncludeDirs, COREPACK_HOME)
	}

	runtimeAptPackages := []string{}
	if p.usesPuppeteer() {
		ctx.Logger.LogInfo("Installing puppeteer dependencies")
		runtimeAptPackages = append(runtimeAptPackages, "xvfb", "gconf-service", "libasound2", "libatk1.0-0", "libc6", "libcairo2", "libcups2", "libdbus-1-3", "libexpat1", "libfontconfig1", "libgbm1", "libgcc1", "libgconf-2-4", "libgdk-pixbuf2.0-0", "libglib2.0-0", "libgtk-3-0", "libnspr4", "libpango-1.0-0", "libpangocairo-1.0-0", "libstdc++6", "libx11-6", "libx11-xcb1", "libxcb1", "libxcomposite1", "libxcursor1", "libxdamage1", "libxext6", "libxfixes3", "libxi6", "libxrandr2", "libxrender1", "libxss1", "libxtst6", "ca-certificates", "fonts-liberation", "libappindicator1", "libnss3", "lsb-release", "xdg-utils", "wget")
	}

	nodeModulesLayer := plan.NewStepLayer(build.Name(), plan.Filter{
		Include: p.packageManager.GetInstallFolder(ctx),
	})
	if p.shouldPrune(ctx) {
		nodeModulesLayer = plan.NewStepLayer(prune.Name(), plan.Filter{
			Include: p.packageManager.GetInstallFolder(ctx),
		})
	}

	buildLayer := plan.NewStepLayer(build.Name(), plan.Filter{
		Include: buildIncludeDirs,
		Exclude: []string{"node_modules", ".yarn"},
	})

	ctx.Deploy.AddAptPackages(runtimeAptPackages)
	ctx.Deploy.AddInputs([]plan.Layer{
		miseStep.GetLayer(),
		nodeModulesLayer,
		buildLayer,
	})

	return nil
}

func (p *NodeProvider) StartCommandHelp() string {
	return "To configure your start command, Railpack will check:\n\n" +
		"1. A \"start\" script in your package.json:\n" +
		"   \"scripts\": {\n" +
		"     \"start\": \"node index.js\"\n" +
		"   }\n\n" +
		"2. A \"main\" field in your package.json pointing to your entry file:\n" +
		"   \"main\": \"src/server.js\"\n\n" +
		"3. An index.js or index.ts file in your project root\n\n" +
		"If you have a static site, you can set the RAILPACK_SPA_OUTPUT_DIR environment variable\n" +
		"containing the directory of your built static files."
}

func (p *NodeProvider) GetStartCommand(ctx *generate.GenerateContext) string {
	if start := p.getScripts(p.packageJson, "start"); start != "" {
		return p.packageManager.RunCmd("start")
	} else if main := p.packageJson.Main; main != "" {
		return p.packageManager.RunScriptCommand(main)
	} else if files, err := ctx.App.FindFiles("{index.js,index.ts}"); err == nil && len(files) > 0 {
		return p.packageManager.RunScriptCommand(files[0])
	} else if p.isNuxt() {
		// Default Nuxt start command
		return "node .output/server/index.mjs"
	}

	return ""
}

func (p *NodeProvider) Build(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	build.AddCommand(plan.NewCopyCommand("."))

	_, ok := p.packageJson.Scripts["build"]
	if ok {
		build.AddCommands([]plan.Command{
			plan.NewExecCommand(p.packageManager.RunCmd("build")),
		})

		if p.isNext() {
			build.AddVariables(map[string]string{"NEXT_TELEMETRY_DISABLED": "1"})
		}
	}

	p.addCaches(ctx, build)
}

func (p *NodeProvider) addFrameworkCaches(ctx *generate.GenerateContext, build *generate.CommandStepBuilder, frameworkName string, frameworkCheck func(*WorkspacePackage, *generate.GenerateContext) bool, cacheSubPath string) {
	if packages, err := p.getPackagesWithFramework(ctx, frameworkCheck); err == nil {
		for _, pkg := range packages {
			var cacheName string
			if pkg.Path == "" {
				cacheName = frameworkName
			} else {
				cacheName = fmt.Sprintf("%s-%s", frameworkName, strings.ReplaceAll(strings.TrimSuffix(pkg.Path, "/"), "/", "-"))
			}
			cacheDir := path.Join("/app", pkg.Path, cacheSubPath)
			build.AddCache(ctx.Caches.AddCache(cacheName, cacheDir))
		}
	}
}

func (p *NodeProvider) addCaches(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	build.AddCache(ctx.Caches.AddCache("node-modules", "/app/node_modules/.cache"))

	p.addFrameworkCaches(ctx, build, "next", func(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
		if pkg.PackageJson.HasScript("build") {
			return strings.Contains(pkg.PackageJson.Scripts["build"], "next build")
		}
		return false
	}, ".next/cache")

	p.addFrameworkCaches(ctx, build, "remix", func(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
		return pkg.PackageJson.hasDependency("@remix-run/node")
	}, ".cache")

	p.addFrameworkCaches(ctx, build, "vite", func(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
		return p.isVitePackage(pkg, ctx)
	}, "node_modules/.vite")

	p.addFrameworkCaches(ctx, build, "astro", func(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
		return p.isAstroPackage(pkg, ctx)
	}, "node_modules/.astro")
}

func (p *NodeProvider) shouldPrune(ctx *generate.GenerateContext) bool {
	return ctx.Env.IsConfigVariableTruthy("PRUNE_DEPS")
}

func (p *NodeProvider) PruneNodeDeps(ctx *generate.GenerateContext, prune *generate.CommandStepBuilder) {
	ctx.Logger.LogInfo("Pruning node dependencies")
	prune.Variables["NPM_CONFIG_PRODUCTION"] = "true"
	prune.Secrets = []string{}
	p.packageManager.PruneDeps(ctx, prune)
}

func (p *NodeProvider) InstallNodeDeps(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) {
	maps.Copy(install.Variables, p.GetNodeEnvVars(ctx))
	install.Secrets = []string{}
	install.UseSecretsWithPrefixes([]string{"NODE", "NPM", "BUN", "PNPM", "YARN", "CI"})
	install.AddPaths([]string{"/app/node_modules/.bin"})

	if ctx.App.HasMatch("node_modules") {
		ctx.Logger.LogWarn("node_modules directory found in project root, this is likely a mistake")
		ctx.Logger.LogWarn("It is recommended to add node_modules to the .gitignore file")
	}

	if p.usesCorepack() {
		pmName, pmVersion := p.packageJson.GetPackageManagerInfo()
		install.AddVariables(map[string]string{
			"COREPACK_HOME": COREPACK_HOME,
		})
		ctx.Logger.LogInfo("Installing %s@%s with Corepack", pmName, pmVersion)

		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("package.json"),
			plan.NewExecShellCommand("npm i -g corepack@latest && corepack enable && corepack prepare --activate"),
		})
	}

	p.packageManager.installDependencies(ctx, p.workspace, install)
}

func (p *NodeProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	requiresNode := p.requiresNode(ctx)

	// Node
	if requiresNode {
		node := miseStep.Default("node", DEFAULT_NODE_VERSION)

		if envVersion, varName := ctx.Env.GetConfigVariable("NODE_VERSION"); envVersion != "" {
			miseStep.Version(node, envVersion, varName)
		}

		if p.packageJson != nil && p.packageJson.Engines != nil && p.packageJson.Engines["node"] != "" {
			miseStep.Version(node, p.packageJson.Engines["node"], "package.json > engines > node")
		}

		if nvmrc, err := ctx.App.ReadFile(".nvmrc"); err == nil {
			if len(nvmrc) > 0 && nvmrc[0] == 'v' {
				nvmrc = nvmrc[1:]
			}

			miseStep.Version(node, string(nvmrc), ".nvmrc")
		}

		if nodeVersionFile, err := ctx.App.ReadFile(".node-version"); err == nil {
			miseStep.Version(node, string(nodeVersionFile), ".node-version")
		}
	}

	// Bun
	if p.requiresBun(ctx) {
		bun := miseStep.Default("bun", DEFAULT_BUN_VERSION)

		if envVersion, varName := ctx.Env.GetConfigVariable("BUN_VERSION"); envVersion != "" {
			miseStep.Version(bun, envVersion, varName)
		}

		// If we don't need node in the final image, we still want to include it for the install steps
		// since many packages need node-gyp to install native modules
		// in this case, we don't need a specific version, so we'll just pull from apt
		if !requiresNode && ctx.Config.Packages["node"] == "" {
			miseStep.AddSupportingAptPackage("nodejs")
		}
	}

	p.packageManager.GetPackageManagerPackages(ctx, p.packageJson, miseStep)

	if p.usesCorepack() {
		miseStep.Variables["MISE_NODE_COREPACK"] = "true"
	}
}

func (p *NodeProvider) GetNodeEnvVars(ctx *generate.GenerateContext) map[string]string {
	envVars := map[string]string{
		"NODE_ENV":                   "production",
		"NPM_CONFIG_PRODUCTION":      "false",
		"NPM_CONFIG_UPDATE_NOTIFIER": "false",
		"NPM_CONFIG_FUND":            "false",
		"CI":                         "true",
	}

	if p.packageManager == PackageManagerYarn1 {
		envVars["YARN_PRODUCTION"] = "false"
		envVars["MISE_YARN_SKIP_GPG"] = "true" // https://github.com/mise-plugins/mise-yarn/pull/8
	}

	if p.isAstro(ctx) && !p.isAstroSPA(ctx) {
		maps.Copy(envVars, p.getAstroEnvVars())
	}

	return envVars
}

func (p *NodeProvider) hasDependency(dependency string) bool {
	return p.packageJson.hasDependency(dependency)
}

func (p *NodeProvider) usesCorepack() bool {
	return p.packageJson != nil && p.packageJson.PackageManager != nil && p.packageManager != PackageManagerBun
}

func (p *NodeProvider) usesPuppeteer() bool {
	return p.workspace.HasDependency("puppeteer")
}

func (p *NodeProvider) getPackageManager(app *app.App) PackageManager {
	packageManager := PackageManagerNpm

	// Check packageManager field first
	if packageJson, err := p.GetPackageJson(app); err == nil && packageJson.PackageManager != nil {
		pmName, pmVersion := packageJson.GetPackageManagerInfo()
		if pmName == "yarn" && pmVersion != "" {
			majorVersion := strings.Split(pmVersion, ".")[0]
			if majorVersion == "1" {
				return PackageManagerYarn1
			} else {
				return PackageManagerYarnBerry
			}
		} else if pmName == "pnpm" {
			return PackageManagerPnpm
		} else if pmName == "bun" {
			return PackageManagerBun
		}
	}

	// Fall back to file-based detection
	if app.HasMatch("pnpm-lock.yaml") {
		packageManager = PackageManagerPnpm
	} else if app.HasMatch("bun.lockb") || app.HasMatch("bun.lock") {
		packageManager = PackageManagerBun
	} else if app.HasMatch(".yarnrc.yml") || app.HasMatch(".yarnrc.yaml") {
		packageManager = PackageManagerYarnBerry
	} else if app.HasMatch("yarn.lock") {
		packageManager = PackageManagerYarn1
	}

	return packageManager
}

func (p *NodeProvider) GetPackageJson(app *app.App) (*PackageJson, error) {
	packageJson := NewPackageJson()
	if !app.HasMatch("package.json") {
		return packageJson, nil
	}

	err := app.ReadJSON("package.json", packageJson)
	if err != nil {
		return nil, fmt.Errorf("error reading package.json: %w", err)
	}

	return packageJson, nil
}

func (p *NodeProvider) getScripts(packageJson *PackageJson, name string) string {
	if scripts := packageJson.Scripts; scripts != nil {
		if script, ok := scripts[name]; ok {
			return script
		}
	}

	return ""
}

func (p *NodeProvider) SetNodeMetadata(ctx *generate.GenerateContext) {
	runtime := p.getRuntime(ctx)
	spaFramework := p.getSPAFramework(ctx)

	ctx.Metadata.Set("nodeRuntime", runtime)
	ctx.Metadata.Set("nodeSPAFramework", spaFramework)
	ctx.Metadata.Set("nodePackageManager", string(p.packageManager))
	ctx.Metadata.SetBool("nodeIsSPA", p.isSPA(ctx))
	ctx.Metadata.SetBool("nodeUsesCorepack", p.usesCorepack())
}

func (p *NodeProvider) getPackagesWithFramework(ctx *generate.GenerateContext, frameworkCheck func(*WorkspacePackage, *generate.GenerateContext) bool) ([]*WorkspacePackage, error) {
	packages := []*WorkspacePackage{}

	// Check root package first
	if p.workspace != nil && frameworkCheck(p.workspace.Root, ctx) {
		packages = append(packages, p.workspace.Root)
	}

	// Check workspace packages
	if p.workspace != nil {
		for _, pkg := range p.workspace.Packages {
			if frameworkCheck(pkg, ctx) {
				packages = append(packages, pkg)
			}
		}
	}

	return packages, nil
}

func (p *NodeProvider) requiresNode(ctx *generate.GenerateContext) bool {
	if p.packageManager != PackageManagerBun || p.packageJson == nil || p.packageJson.PackageManager != nil {
		return true
	}

	for _, script := range p.packageJson.Scripts {
		if strings.Contains(script, "node") {
			return true
		}
	}

	return p.isAstro(ctx)
}

// packageJsonRequiresBun checks if a package.json's scripts use bun commands
func packageJsonRequiresBun(packageJson *PackageJson) bool {
	if packageJson == nil || packageJson.Scripts == nil {
		return false
	}

	for _, script := range packageJson.Scripts {
		if bunCommandRegex.MatchString(script) {
			return true
		}
	}

	return false
}

// requiresBun checks if bun should be installed and available for the build and final image
func (p *NodeProvider) requiresBun(ctx *generate.GenerateContext) bool {
	if p.packageManager == PackageManagerBun {
		return true
	}

	if packageJsonRequiresBun(p.packageJson) {
		return true
	}

	if ctx.Config.Deploy != nil && bunCommandRegex.MatchString(ctx.Config.Deploy.StartCmd) {
		return true
	}

	return false
}

func (p *NodeProvider) getRuntime(ctx *generate.GenerateContext) string {
	if p.isSPA(ctx) {
		if p.isAstro(ctx) {
			return "astro"
		} else if p.isVite(ctx) {
			return "vite"
		} else if p.isCRA(ctx) {
			return "cra"
		} else if p.isAngular(ctx) {
			return "angular"
		} else if p.isReactRouter(ctx) {
			return "react-router"
		}

		return "static"
	} else if p.isNext() {
		return "next"
	} else if p.isNuxt() {
		return "nuxt"
	} else if p.isRemix() {
		return "remix"
	} else if p.isTanstackStart() {
		return "tanstack-start"
	} else if p.isVite(ctx) {
		return "vite"
	} else if p.isReactRouter(ctx) {
		return "react-router"
	} else if p.packageManager == PackageManagerBun {
		return "bun"
	}

	return "node"
}

func (p *NodeProvider) isNext() bool {
	return p.hasDependency("next")
}

func (p *NodeProvider) isNuxt() bool {
	return p.hasDependency("nuxt")
}

func (p *NodeProvider) isRemix() bool {
	return p.hasDependency("@remix-run/node")
}

func (p *NodeProvider) isTanstackStart() bool {
	return p.hasDependency("@tanstack/react-start")
}
