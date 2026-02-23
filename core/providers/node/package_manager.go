package node

import (
	"fmt"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	PackageManagerNpm       PackageManager = "npm"
	PackageManagerPnpm      PackageManager = "pnpm"
	PackageManagerBun       PackageManager = "bun"
	PackageManagerYarn1     PackageManager = "yarn1"
	PackageManagerYarnBerry PackageManager = "yarnberry"

	DEFAULT_PNPM_VERSION = "9"
)

func (p PackageManager) Name() string {
	switch p {
	case PackageManagerNpm:
		return "npm"
	case PackageManagerPnpm:
		return "pnpm"
	case PackageManagerBun:
		return "bun"
	case PackageManagerYarn1, PackageManagerYarnBerry:
		return "yarn"
	default:
		log.Warnf("unknown package manager: %s", p)
		return ""
	}
}

func (p PackageManager) RunCmd(cmd string) string {
	return fmt.Sprintf("%s run %s", p.Name(), cmd)
}

func (p PackageManager) RunScriptCommand(cmd string) string {
	if p == PackageManagerBun {
		return "bun " + cmd
	}
	return "node " + cmd
}

func (p PackageManager) installDependencies(ctx *generate.GenerateContext, workspace *Workspace, install *generate.CommandStepBuilder, usingCorepack bool) {
	packageJsons := workspace.AllPackageJson()

	hasPreInstall := false
	hasPostInstall := false
	hasPrepare := false

	for _, packageJson := range packageJsons {
		hasPreInstall = hasPreInstall || (packageJson.Scripts != nil && packageJson.Scripts["preinstall"] != "")
		hasPostInstall = hasPostInstall || (packageJson.Scripts != nil && packageJson.Scripts["postinstall"] != "")
		hasPrepare = hasPrepare || (packageJson.Scripts != nil && packageJson.Scripts["prepare"] != "")
	}

	usesLocalFile := p.usesLocalFile(ctx)

	// If there are any pre/post install scripts, we need the entire app to be copied
	// This is to handle things like patch-package
	if hasPreInstall || hasPostInstall || hasPrepare || usesLocalFile {
		layer := ctx.NewLocalLayer()
		layer.Filter.Exclude = append(layer.Filter.Exclude, "node_modules")
		install.AddInput(layer)

		// Use all secrets for the install step if there are any pre/post install scripts
		install.UseSecrets([]string{"*"})
	} else {
		files := p.SupportingInstallFiles(ctx)
		for _, file := range files {
			install.AddCommands([]plan.Command{
				plan.NewCopyCommand(file, file),
			})
		}
	}

	p.installDeps(ctx, install, usingCorepack)
}

// GetCache returns the cache for the package manager
func (p PackageManager) GetInstallCache(ctx *generate.GenerateContext) string {
	switch p {
	case PackageManagerNpm:
		return ctx.Caches.AddCache("npm-install", "/root/.npm")
	case PackageManagerPnpm:
		return ctx.Caches.AddCache("pnpm-install", "/root/.local/share/pnpm/store/v3")
	case PackageManagerBun:
		return ctx.Caches.AddCache("bun-install", "/root/.bun/install/cache")
	case PackageManagerYarn1:
		return ctx.Caches.AddCacheWithType("yarn-install", "/usr/local/share/.cache/yarn", plan.CacheTypeLocked)
	case PackageManagerYarnBerry:
		return ctx.Caches.AddCache("yarn-install", "/app/.yarn/cache")
	default:
		return ""
	}
}

func (p PackageManager) installDeps(ctx *generate.GenerateContext, install *generate.CommandStepBuilder, usingCorepack bool) {
	install.AddCache(p.GetInstallCache(ctx))

	switch p {
	case PackageManagerNpm:
		hasLockfile := ctx.App.HasFile("package-lock.json")
		if hasLockfile {
			install.AddCommand(plan.NewExecCommand("npm ci"))
		} else {
			install.AddCommand(plan.NewExecCommand("npm install"))
		}
	case PackageManagerPnpm:
		// pnpm (standalone) does not bundle node-gyp like npm does, so we must install it globally
		// to support packages with native dependencies (e.g., better-sqlite3, bcrypt, etc.)
		// Only needed when using mise to install pnpm (not corepack, which includes node-gyp)
		if !usingCorepack {
			// Set PNPM_HOME so pnpm can create a global bin directory for node-gyp
			install.AddEnvVars(map[string]string{
				"PNPM_HOME": "/pnpm",
			})
			install.AddPaths([]string{"/pnpm"})
			install.AddCommand(plan.NewExecCommand("pnpm add -g node-gyp"))
		}

		hasLockfile := ctx.App.HasFile("pnpm-lock.yaml")
		if hasLockfile {
			install.AddCommand(plan.NewExecCommand("pnpm install --frozen-lockfile --prefer-offline"))
		} else {
			install.AddCommand(plan.NewExecCommand("pnpm install"))
		}
	case PackageManagerBun:
		install.AddCommand(plan.NewExecCommand("bun install --frozen-lockfile"))
	case PackageManagerYarn1:
		install.AddCommand(plan.NewExecCommand("yarn install --frozen-lockfile"))
	case PackageManagerYarnBerry:
		install.AddCommand(plan.NewExecCommand("yarn install --check-cache"))
	}
}

func (p PackageManager) PruneDeps(ctx *generate.GenerateContext, prune *generate.CommandStepBuilder) {
	prune.AddCache(p.GetInstallCache(ctx))

	if pruneCmd, _ := ctx.Env.GetConfigVariable("NODE_PRUNE_CMD"); pruneCmd != "" {
		prune.AddCommand(plan.NewExecCommand(pruneCmd))
		return
	}

	switch p {
	case PackageManagerNpm:
		prune.AddCommand(plan.NewExecCommand("npm prune --omit=dev --ignore-scripts"))
	case PackageManagerPnpm:
		p.prunePnpm(ctx, prune)
	case PackageManagerBun:
		// Prune is not supported in Bun. https://github.com/oven-sh/bun/issues/3605
		prune.AddCommand(plan.NewExecShellCommand("rm -rf node_modules && bun install --production --ignore-scripts"))
	case PackageManagerYarn1:
		prune.AddCommand(plan.NewExecCommand("yarn install --production=true"))
	case PackageManagerYarnBerry:
		p.pruneYarnBerry(ctx, prune)
	}
}

func (p PackageManager) prunePnpm(ctx *generate.GenerateContext, prune *generate.CommandStepBuilder) {
	if packageJson, err := p.getPackageJsonFromContext(ctx); err == nil {
		_, pnpmVersion := packageJson.GetPackageManagerInfo()
		if pnpmVersion != "" {
			pnpmVersion, err := semver.NewVersion(pnpmVersion)

			// pnpm 8.15.6 added the --ignore-scripts flag to the prune command
			// https://github.com/pnpm/pnpm/releases/tag/v8.15.6
			if err == nil && pnpmVersion.Compare(semver.MustParse("8.15.6")) == -1 {
				prune.AddCommand(plan.NewExecCommand("pnpm prune --prod"))
				return
			}
		}
	}

	prune.AddCommand(plan.NewExecCommand("pnpm prune --prod --ignore-scripts"))
}

func (p PackageManager) pruneYarnBerry(ctx *generate.GenerateContext, prune *generate.CommandStepBuilder) {
	// Check if we can determine the Yarn version from packageManager field
	if packageJson, err := p.getPackageJsonFromContext(ctx); err == nil {
		_, version := packageJson.GetPackageManagerInfo()
		if version != "" && strings.HasPrefix(version, "3.") {
			// If you know of the proper way to prune Yarn 3, please make a PR
			ctx.Logger.LogWarn("Yarn 3 doesn't have a prune command, using install instead")
			prune.AddCommand(plan.NewExecCommand("yarn install --check-cache"))
			return
		}
	}

	// Yarn 2 and 4+ support workspaces focus (also fallback for unknown versions)
	// Note: yarn workspaces focus doesn't support --ignore-scripts flag
	prune.AddCommand(plan.NewExecCommand("yarn workspaces focus --production --all"))
}

func (p PackageManager) getPackageJsonFromContext(ctx *generate.GenerateContext) (*PackageJson, error) {
	packageJson := NewPackageJson()
	if !ctx.App.HasFile("package.json") {
		return packageJson, nil
	}

	err := ctx.App.ReadJSON("package.json", packageJson)
	if err != nil {
		return nil, err
	}

	return packageJson, nil
}

func (p PackageManager) GetInstallFolder(ctx *generate.GenerateContext) []string {
	switch p {
	case PackageManagerYarnBerry:
		installFolders := []string{"/app/.yarn", p.getYarnBerryGlobalFolder(ctx)}
		if p.getYarnBerryNodeLinker(ctx) == "node-modules" {
			installFolders = append(installFolders, "/app/node_modules")
		}
		return installFolders
	default:
		return []string{"/app/node_modules"}
	}
}

// SupportingInstallFiles returns a list of files that are needed to install dependencies
func (p PackageManager) SupportingInstallFiles(ctx *generate.GenerateContext) []string {
	// Use brace expansion for single filesystem traversal instead of 16 separate globs
	pattern := "**/{package.json,package-lock.json,pnpm-workspace.yaml,yarn.lock,pnpm-lock.yaml,bun.lockb,bun.lock,.yarn,.pnp.*,.yarnrc.yml,.npmrc,.node-version,.nvmrc,patches,.pnpm-patches,prisma}"

	var allFiles []string

	files, err := ctx.App.FindFiles(pattern)
	if err == nil {
		for _, file := range files {
			if !strings.HasPrefix(file, "node_modules/") {
				allFiles = append(allFiles, file)
			}
		}
	}

	dirs, err := ctx.App.FindDirectories(pattern)
	if err == nil {
		allFiles = append(allFiles, dirs...)
	}

	if customInstallPatterns, _ := ctx.Env.GetConfigVariableList("NODE_INSTALL_PATTERNS"); len(customInstallPatterns) > 0 {
		ctx.Logger.LogInfo("Using custom install patterns: %s", strings.Join(customInstallPatterns, " "))
		for _, pat := range customInstallPatterns {
			customFiles, _ := ctx.App.FindFiles("**/" + pat)
			allFiles = append(allFiles, customFiles...)
		}
	}

	return allFiles
}

// GetPackageManagerPackages installs specific versions of package managers by analyzing the users code
func (p PackageManager) GetPackageManagerPackages(ctx *generate.GenerateContext, packageJson *PackageJson, packages *generate.MiseStepBuilder) {
	pmName, pmVersion := packageJson.GetPackageManagerInfo()

	// Pnpm
	if p == PackageManagerPnpm {
		pnpm := packages.Default("pnpm", DEFAULT_PNPM_VERSION)

		// Prefer explicit version from package.json engines over defaults/lockfile
		if packageJson != nil && packageJson.Engines != nil && packageJson.Engines["pnpm"] != "" {
			packages.Version(pnpm, packageJson.Engines["pnpm"], "package.json > engines > pnpm")
		}

		lockfile, err := ctx.App.ReadFile("pnpm-lock.yaml")
		if err == nil {
			if strings.HasPrefix(lockfile, "lockfileVersion: 5.3") {
				packages.Version(pnpm, "6", "pnpm-lock.yaml")
			} else if strings.HasPrefix(lockfile, "lockfileVersion: 5.4") {
				packages.Version(pnpm, "7", "pnpm-lock.yaml")
			} else if strings.HasPrefix(lockfile, "lockfileVersion: '6.0'") {
				packages.Version(pnpm, "8", "pnpm-lock.yaml")
			}
		}

		if pmName == "pnpm" && pmVersion != "" {
			packages.Version(pnpm, pmVersion, "package.json > packageManager")

			// skip installing via Mise and install with corepack instead
			// https://github.com/railwayapp/railpack/issues/201
			packages.SkipMiseInstall(pnpm)
		}
	}

	// Yarn
	if p == PackageManagerYarn1 || p == PackageManagerYarnBerry {
		var defaultMajor string
		if p == PackageManagerYarn1 {
			defaultMajor = "1"
			packages.AddSupportingAptPackage("tar")
			packages.AddSupportingAptPackage("gpg")
		} else {
			defaultMajor = "2"
		}
		yarn := packages.Default("yarn", defaultMajor)

		// Prefer explicit version from package.json engines over defaults
		if packageJson != nil && packageJson.Engines != nil && packageJson.Engines["yarn"] != "" {
			packages.Version(yarn, packageJson.Engines["yarn"], "package.json > engines > yarn")
		}

		// TODO we should use SemVer at this point
		if pmName == "yarn" && pmVersion != "" {
			majorVersion := strings.Split(pmVersion, ".")[0]
			yarn := packages.Default("yarn", majorVersion)
			packages.Version(yarn, pmVersion, "package.json > packageManager")

			// skip installing via Mise and install with corepack instead
			// https://github.com/railwayapp/railpack/issues/201
			packages.SkipMiseInstall(yarn)
		}
	}

	// Bun
	if p == PackageManagerBun {
		bun := packages.Default("bun", "latest")

		// Prefer explicit version from package.json engines over defaults
		if packageJson != nil && packageJson.Engines != nil && packageJson.Engines["bun"] != "" {
			packages.Version(bun, packageJson.Engines["bun"], "package.json > engines > bun")
		}

		if pmName == "bun" && pmVersion != "" {
			packages.Version(bun, pmVersion, "package.json > packageManager")
		}
	}
}

// usesLocalFile returns true if the package.json has a local dependency (e.g. file:./path/to/package)
func (p PackageManager) usesLocalFile(ctx *generate.GenerateContext) bool {
	files, err := ctx.App.FindFiles("**/package.json")
	if err != nil {
		return false
	}

	for _, file := range files {
		packageJson := &PackageJson{}
		err := ctx.App.ReadJSON(file, packageJson)
		if err != nil {
			continue
		}

		if packageJson.hasLocalDependency() {
			return true
		}
	}

	return false
}

type YarnRc struct {
	GlobalFolder string `yaml:"globalFolder"`
	NodeLinker   string `yaml:"nodeLinker"`
}

func (p PackageManager) getYarnRc(ctx *generate.GenerateContext) YarnRc {
	var yarnRc YarnRc
	if err := ctx.App.ReadYAML(".yarnrc.yml", &yarnRc); err == nil {
		return yarnRc
	}
	return YarnRc{}
}

func (p PackageManager) getYarnBerryGlobalFolder(ctx *generate.GenerateContext) string {
	yarnRc := p.getYarnRc(ctx)
	if yarnRc.GlobalFolder != "" {
		return yarnRc.GlobalFolder
	}

	return "/root/.yarn"
}

func (p PackageManager) getYarnBerryNodeLinker(ctx *generate.GenerateContext) string {
	yarnRc := p.getYarnRc(ctx)
	if yarnRc.NodeLinker != "" {
		return yarnRc.NodeLinker
	}
	return "pnp"
}
