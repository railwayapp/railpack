package node

import (
	"fmt"
	"maps"
	"path"
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

	if p.packageManager == PackageManagerYarn2 {
		buildIncludeDirs = append(buildIncludeDirs, p.packageManager.getYarn2GlobalFolder(ctx))
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

func (p *NodeProvider) addCaches(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	build.AddCache(ctx.Caches.AddCache("node-modules", "/app/node_modules/.cache"))

	if nextApps, err := p.getNextApps(ctx); err == nil {
		for _, nextApp := range nextApps {
			nextCacheDir := path.Join("/app", nextApp, ".next/cache")
			build.AddCache(ctx.Caches.AddCache(fmt.Sprintf("next-%s", nextApp), nextCacheDir))
		}
	}

	if p.isRemix() {
		build.AddCache(ctx.Caches.AddCache("remix", ".cache"))
	}

	if p.isAstro(ctx) {
		build.AddCache(p.getAstroCache(ctx))
	}

	if p.isVite(ctx) {
		build.AddCache(p.getViteCache(ctx))
	}

	if p.isNuxt() {
		build.AddCache(ctx.Caches.AddCache("nuxt", ".nuxt"))
	}
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

	if p.usesCorepack() {
		install.AddVariables(map[string]string{
			"COREPACK_HOME": COREPACK_HOME,
		})
		ctx.Logger.LogInfo("Using Corepack")

		install.AddCommands([]plan.Command{
			plan.NewCopyCommand("package.json"),
			plan.NewExecShellCommand("npm i -g corepack@latest && corepack enable && corepack prepare --activate"),
		})
	}

	p.packageManager.installDependencies(ctx, p.packageJson, install)
}

func (p *NodeProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	requiresNode := p.requiresNode(ctx)

	// Node
	if requiresNode {
		node := miseStep.Default("node", DEFAULT_NODE_VERSION)

		if envVersion, varName := ctx.Env.GetConfigVariable("NODE_VERSION"); envVersion != "" {
			miseStep.Version(node, envVersion, varName)
		}

		if p.packageJson.Engines != nil && p.packageJson.Engines["node"] != "" {
			miseStep.Version(node, p.packageJson.Engines["node"], "package.json > engines > node")
		}

		if nvmrc, err := ctx.App.ReadFile(".nvmrc"); err == nil {
			if len(nvmrc) > 0 && nvmrc[0] == 'v' {
				nvmrc = nvmrc[1:]
			}

			miseStep.Version(node, string(nvmrc), ".nvmrc")
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
	return p.packageJson.PackageManager != nil && p.packageManager != PackageManagerBun
}

func (p *NodeProvider) usesPuppeteer() bool {
	return p.workspace.HasDependency("puppeteer")
}

func (p *NodeProvider) getPackageManager(app *app.App) PackageManager {
	packageManager := PackageManagerNpm

	if app.HasMatch("pnpm-lock.yaml") {
		packageManager = PackageManagerPnpm
	} else if app.HasMatch("bun.lockb") || app.HasMatch("bun.lock") {
		packageManager = PackageManagerBun
	} else if app.HasMatch(".yarnrc.yml") || app.HasMatch(".yarnrc.yaml") {
		packageManager = PackageManagerYarn2
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

func (p *NodeProvider) getNextApps(ctx *generate.GenerateContext) ([]string, error) {
	nextPaths, err := p.filterPackageJson(ctx, func(packageJson *PackageJson) bool {
		if packageJson.HasScript("build") {
			return strings.Contains(packageJson.Scripts["build"], "next build")
		}

		return false
	})
	if err != nil {
		return nil, err
	}

	return nextPaths, nil
}

func (p *NodeProvider) filterPackageJson(ctx *generate.GenerateContext, filterFunc func(packageJson *PackageJson) bool) ([]string, error) {
	filteredPaths := []string{}

	files, err := ctx.App.FindFiles("**/package.json")
	if err != nil {
		return filteredPaths, err
	}

	for _, file := range files {
		var packageJson PackageJson
		err := ctx.App.ReadJSON(file, &packageJson)
		if err != nil {
			return filteredPaths, err
		}

		if filterFunc(&packageJson) {
			dirPath := strings.TrimSuffix(file, "package.json")
			filteredPaths = append(filteredPaths, dirPath)
		}
	}

	return filteredPaths, nil
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

func (p *NodeProvider) requiresBun(ctx *generate.GenerateContext) bool {
	if p.packageManager == PackageManagerBun {
		return true
	}

	for _, script := range p.packageJson.Scripts {
		if strings.Contains(script, "bun") {
			return true
		}
	}

	if ctx.Config.Deploy != nil && strings.Contains(ctx.Config.Deploy.StartCmd, "bun") {
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
		}

		return "static"
	} else if p.isNext() {
		return "next"
	} else if p.isNuxt() {
		return "nuxt"
	} else if p.isRemix() {
		return "remix"
	} else if p.isVite(ctx) {
		return "vite"
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
