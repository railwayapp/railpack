package node

import (
	"fmt"
	"maps"
	"path"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
)

type PackageManager string

const (
	DEFAULT_NODE_VERSION = "22"
	DEFAULT_BUN_VERSION  = "latest"

	COREPACK_HOME = "/opt/corepack"

	// not used by npm, but many other tools: next, jest, webpack, etc
	NODE_MODULES_CACHE = "/app/node_modules/.cache"
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
	return ctx.App.HasFile("package.json"), nil
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

	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepLayer(miseStep.Name()))
	p.InstallNodeDeps(ctx, install)

	prune := ctx.NewCommandStep("prune")
	prune.AddInput(plan.NewStepLayer(install.Name()))
	prune.Secrets = []string{}
	if p.shouldPrune(ctx) && !isSPA {
		p.PruneNodeDeps(ctx, prune)
	}

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
		// TODO we should just have a default dockerignore/exclusion list instead of hardcoding here
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
	build.AddInput(ctx.NewLocalLayer())

	_, ok := p.packageJson.Scripts["build"]
	if ok {
		build.AddCommands([]plan.Command{
			plan.NewExecCommand(p.packageManager.RunCmd("build")),
		})

		if p.isNext() {
			build.AddVariables(map[string]string{"NEXT_TELEMETRY_DISABLED": "1"})
		}
	}

	p.addCachesToBuildStep(ctx, build)
}

// adds framework-specific caches for packages that match the given framework check.
// It creates uniquely named caches for each package's framework subdirectory to optimize build performance.
func (p *NodeProvider) addFrameworkCaches(ctx *generate.GenerateContext, build *generate.CommandStepBuilder, frameworkName string, frameworkCheck func(*WorkspacePackage, *generate.GenerateContext) bool, cacheSubPath string) {
	if packages, err := p.getPackagesWithFramework(ctx, frameworkCheck); err == nil {
		for _, pkg := range packages {
			var cacheName string
			if pkg.Path == "" {
				cacheName = frameworkName
			} else {
				cacheName = fmt.Sprintf("%s-%s", frameworkName, strings.ReplaceAll(strings.TrimSuffix(pkg.Path, "/"), "/", "-"))
			}
			// in this case, pkg.Path represents the relative path to a workspace package from the root of your repository
			cacheDir := path.Join("/app", pkg.Path, cacheSubPath)
			build.AddCache(ctx.Caches.AddCache(cacheName, cacheDir))
		}
	}
}

// cache directories to add to the build step: if lock files are unchanged, these are pulled from cache, but cannot
// be removed in future steps.
func (p *NodeProvider) addCachesToBuildStep(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	build.AddCache(ctx.Caches.AddCache("node-modules", NODE_MODULES_CACHE))

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

	p.addFrameworkCaches(ctx, build, "react-router", func(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
		return p.isReactRouterPackage(pkg, ctx)
	}, ".react-router")
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

	// TODO once dockerignore is in place, we should remove this
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
			// corepack will detect the package manager version from package.json, safe to assume the user is properly
			// specifying the version they want there, no need to check other version specifications.
			// corepack *used* to be bundled with node, but as of v25 it's not, so we install it explicitly
			plan.NewExecShellCommand("npm i -g corepack@latest && corepack enable && corepack prepare --activate"),
		})
	}
	install.AddCommands([]plan.Command{
		// it's possible for a package.json to exist without any dependencies, in which case node_modules is not generated
		// and bun.lockb, etc are not generated either. However, this path is used to compute the cache key, so we ensure
		// it exists on the filesystem to avoid a docker cache key computation error.
		plan.NewExecCommand(fmt.Sprintf("mkdir -p %s", NODE_MODULES_CACHE)),
	})

	p.packageManager.installDependencies(ctx, p.workspace, install, p.usesCorepack())
}

// resolve node version selection which is used both for node runtime *and* when bun is used but node is required for
// build or runtime.
func (p *NodeProvider) applyNodeVersionResolution(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder, nodeToolRef resolver.PackageRef) {
	if envVersion, varName := ctx.Env.GetConfigVariable("NODE_VERSION"); envVersion != "" {
		miseStep.Version(nodeToolRef, envVersion, varName)
	}

	if p.packageJson != nil && p.packageJson.Engines != nil && p.packageJson.Engines["node"] != "" {
		miseStep.Version(nodeToolRef, p.packageJson.Engines["node"], "package.json > engines > node")
	}

	// TODO both nvmrc and node-version should be parsed via mise idiomatic version parsing
	if nvmrc, err := ctx.App.ReadFile(".nvmrc"); err == nil {
		if len(nvmrc) > 0 && nvmrc[0] == 'v' {
			nvmrc = nvmrc[1:]
		}

		miseStep.Version(nodeToolRef, string(nvmrc), ".nvmrc")
	}

	if nodeVersionFile, err := ctx.App.ReadFile(".node-version"); err == nil {
		miseStep.Version(nodeToolRef, string(nodeVersionFile), ".node-version")
	}
}

func (p *NodeProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	requiresNode := p.requiresNode(ctx)
	misePackages := []string{}

	// Node
	if requiresNode {
		node := miseStep.Default("node", DEFAULT_NODE_VERSION)
		misePackages = append(misePackages, "node")

		// libatomic1 is required for Node.js v25+
		ctx.Deploy.AddAptPackages([]string{"libatomic1"})

		p.applyNodeVersionResolution(ctx, miseStep, node)
	}

	// Bun
	if p.requiresBun(ctx) {
		bun := miseStep.Default("bun", DEFAULT_BUN_VERSION)
		misePackages = append(misePackages, "bun")

		if envVersion, varName := ctx.Env.GetConfigVariable("BUN_VERSION"); envVersion != "" {
			miseStep.Version(bun, envVersion, varName)
		}

		// .bun-version is a community convention for specifying the Bun version.
		// It is not officially supported by Bun itself, but is recognized by version managers like mise.
		if bunVersionFile, err := ctx.App.ReadFile(".bun-version"); err == nil {
			miseStep.Version(bun, string(bunVersionFile), ".bun-version")
		}

		// If we don't need node in the final image, we still want to include it for the install steps
		// since many packages need node-gyp to install native modules
		if !requiresNode && ctx.Config.Packages["node"] == "" {
			node := miseStep.Default("node", DEFAULT_NODE_VERSION)
			misePackages = append(misePackages, "node")

			p.applyNodeVersionResolution(ctx, miseStep, node)

			// libatomic1 is required for Node.js v25+
			ctx.Deploy.AddAptPackages([]string{"libatomic1"})
		}
	}

	p.packageManager.GetPackageManagerPackages(ctx, p.packageJson, miseStep)

	if p.usesCorepack() {
		miseStep.Variables["MISE_NODE_COREPACK"] = "true"
	}

	// Check for mise.toml and .tool-versions and use those versions if they exist
	if len(misePackages) > 0 {
		miseStep.UseMiseVersions(ctx, misePackages)
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
	}

	if p.isAstro(ctx) && !p.isAstroSPA(ctx) {
		maps.Copy(envVars, p.getAstroEnvVars())
	}

	return envVars
}

func (p *NodeProvider) hasDependency(dependency string) bool {
	return p.packageJson.hasDependency(dependency)
}

// if 'packageManager' field exists in package.json, then assume corepack unless using bun
func (p *NodeProvider) usesCorepack() bool {
	return p.packageJson != nil && p.packageJson.PackageManager != nil && p.packageManager != PackageManagerBun
}

func (p *NodeProvider) usesPuppeteer() bool {
	return p.workspace.HasDependency("puppeteer")
}

// determine the major version of yarn from a version string. These major versions are installed and managed quite
// differently which is why we need to distinguish them here.
func parseYarnPackageManager(pmVersion string) PackageManager {
	if strings.Split(pmVersion, ".")[0] == "1" {
		return PackageManagerYarn1
	}

	// versions 2-4 are all considered part of the "Yarn Berry" release line
	return PackageManagerYarnBerry
}

func (p *NodeProvider) getPackageManager(app *app.App) PackageManager {
	// Check packageManager field first
	if packageJson, err := p.GetPackageJson(app); err == nil && packageJson.PackageManager != nil {
		pmName, pmVersion := packageJson.GetPackageManagerInfo()
		if pmName == "yarn" && pmVersion != "" {
			return parseYarnPackageManager(pmVersion)
		} else if pmName == "pnpm" {
			return PackageManagerPnpm
		} else if pmName == "npm" {
			return PackageManagerNpm
		} else if pmName == "bun" {
			return PackageManagerBun
		} else if pmName == "" {
			// this is mostly likely a user configuration bug, so let's at least log it in case someone is stuck
			log.Info("Package manager name is empty in package.json")
		} else {
			log.Warnf("Unknown package manager `%s` specified in package.json, defaulting to npm", pmName)
		}
	}

	// Fall back to file-based detection
	if app.HasFile("pnpm-lock.yaml") {
		return PackageManagerPnpm
	} else if app.HasFile("bun.lockb") || app.HasFile("bun.lock") {
		return PackageManagerBun
	} else if app.HasFile(".yarnrc.yml") || app.HasFile(".yarnrc.yaml") {
		return PackageManagerYarnBerry
	} else if app.HasFile("yarn.lock") {
		return PackageManagerYarn1
	}

	// Finally, consider engines as a last-resort
	if packageJson, err := p.GetPackageJson(app); err == nil && packageJson.Engines != nil {
		if engine := strings.TrimSpace(packageJson.Engines["pnpm"]); engine != "" {
			return PackageManagerPnpm
		}
		if engine := strings.TrimSpace(packageJson.Engines["bun"]); engine != "" {
			return PackageManagerBun
		}
		if engine := strings.TrimSpace(packageJson.Engines["yarn"]); engine != "" {
			// Decide yarn major: 1 -> yarn1, otherwise default to berry
			return parseYarnPackageManager(engine)
		}
	}

	log.Info("No package manager inferred, using npm default")

	return PackageManagerNpm
}

func (p *NodeProvider) GetPackageJson(app *app.App) (*PackageJson, error) {
	packageJson := NewPackageJson()
	if !app.HasFile("package.json") {
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

	return p.isAstro(ctx) || p.isVite(ctx)
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
