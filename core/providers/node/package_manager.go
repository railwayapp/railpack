package node

import (
	"fmt"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	PackageManagerNpm   PackageManager = "npm"
	PackageManagerPnpm  PackageManager = "pnpm"
	PackageManagerBun   PackageManager = "bun"
	PackageManagerYarn1 PackageManager = "yarn1"
	PackageManagerYarn2 PackageManager = "yarn2"

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
	case PackageManagerYarn1, PackageManagerYarn2:
		return "yarn"
	default:
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

func (p PackageManager) installDependencies(ctx *generate.GenerateContext, packageJson *PackageJson, install *generate.CommandStepBuilder) {
	hasPreInstall := packageJson.Scripts != nil && packageJson.Scripts["preinstall"] != ""
	hasPostInstall := packageJson.Scripts != nil && packageJson.Scripts["postinstall"] != ""
	hasPrepare := packageJson.Scripts != nil && packageJson.Scripts["prepare"] != ""
	usesLocalFile := p.usesLocalFile(ctx)

	// If there are any pre/post install scripts, we need the entire app to be copied
	// This is to handle things like patch-package
	if hasPreInstall || hasPostInstall || hasPrepare || usesLocalFile {
		install.AddCommands([]plan.Command{
			plan.NewCopyCommand(".", "."),
		})

		// Use all secrets for the install step if there are any pre/post install scripts
		install.UseSecrets([]string{"*"})
	} else {
		for _, file := range p.SupportingInstallFiles(ctx) {
			install.AddCommands([]plan.Command{
				plan.NewCopyCommand(file, file),
			})
		}
	}

	p.InstallDeps(ctx, install)
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
	case PackageManagerYarn2:
		return ctx.Caches.AddCache("yarn-install", "/app/.yarn/cache")
	default:
		return ""
	}
}

func (p PackageManager) InstallDeps(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) {
	install.AddCache(p.GetInstallCache(ctx))

	switch p {
	case PackageManagerNpm:
		hasLockfile := ctx.App.HasMatch("package-lock.json")
		if hasLockfile {
			install.AddCommand(plan.NewExecCommand("npm ci"))
		} else {
			install.AddCommand(plan.NewExecCommand("npm install"))
		}
	case PackageManagerPnpm:
		install.AddCommand(plan.NewExecCommand("pnpm install --frozen-lockfile --prefer-offline"))
	case PackageManagerBun:
		install.AddCommand(plan.NewExecCommand("bun install --frozen-lockfile"))
	case PackageManagerYarn1:
		install.AddCommand(plan.NewExecCommand("yarn install --frozen-lockfile"))
	case PackageManagerYarn2:
		install.AddCommand(plan.NewExecCommand("yarn install --check-cache"))
	}
}

func (p PackageManager) PruneDeps(ctx *generate.GenerateContext, prune *generate.CommandStepBuilder) {
	prune.AddCache(p.GetInstallCache(ctx))

	switch p {
	case PackageManagerNpm:
		prune.AddCommand(plan.NewExecCommand("npm prune --omit=dev"))
	case PackageManagerPnpm:
		prune.AddCommand(plan.NewExecCommand("pnpm prune --prod"))
	case PackageManagerBun:
		// Prune is not supported in Bun. https://github.com/oven-sh/bun/issues/3605
		prune.AddCommand(plan.NewExecShellCommand("rm -rf node_modules && bun install --production"))
	case PackageManagerYarn1:
		prune.AddCommand(plan.NewExecCommand("yarn install --production=true"))
	case PackageManagerYarn2:
		prune.AddCommand(plan.NewExecCommand("yarn workspaces focus --production --all"))
	}
}

func (p PackageManager) GetInstallFolder(ctx *generate.GenerateContext) []string {
	switch p {
	case PackageManagerYarn2:
		return []string{"/app/.yarn", p.getYarn2GlobalFolder(ctx)}
	default:
		return []string{"/app/node_modules"}
	}
}

// SupportingInstallFiles returns a list of files that are needed to install dependencies
func (p PackageManager) SupportingInstallFiles(ctx *generate.GenerateContext) []string {
	patterns := []string{
		"**/package.json",
		"**/package-lock.json",
		"**/pnpm-workspace.yaml",
		"**/yarn.lock",
		"**/pnpm-lock.yaml",
		"**/bun.lockb",
		"**/bun.lock",
		"**/.yarn",
		"**/.pnp.*",        // Yarn Plug'n'Play files
		"**/.yarnrc.yml",   // Yarn 2+ config
		"**/.npmrc",        // NPM config
		"**/.node-version", // Node version file
		"**/.nvmrc",        // NVM config
		"**/patches",       // PNPM patches
		"**/.pnpm-patches",
		"**/prisma", // To generate Prisma client on install
	}

	if customInstallPatterns, _ := ctx.Env.GetConfigVariable("NODE_INSTALL_PATTERNS"); customInstallPatterns != "" {
		ctx.Logger.LogInfo("Using custom install patterns: %s", customInstallPatterns)
		for _, pattern := range strings.Split(customInstallPatterns, " ") {
			patterns = append(patterns, "**/"+pattern)
		}
	}

	var allFiles []string
	for _, pattern := range patterns {
		files, err := ctx.App.FindFiles(pattern)
		if err != nil {
			continue
		}
		for _, file := range files {
			if !strings.HasPrefix(file, "node_modules/") {
				allFiles = append(allFiles, file)
			}
		}

		dirs, err := ctx.App.FindDirectories(pattern)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, dirs...)
	}

	return allFiles
}

// GetPackageManagerPackages installs specific versions of package managers by analyzing the users code
func (p PackageManager) GetPackageManagerPackages(ctx *generate.GenerateContext, packageJson *PackageJson, packages *generate.MiseStepBuilder) {
	// Pnpm
	if p == PackageManagerPnpm {
		pnpm := packages.Default("pnpm", DEFAULT_PNPM_VERSION)

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

		name, version := p.parsePackageManagerField(packageJson)
		if name == "pnpm" && version != "" {
			packages.Version(pnpm, version, "package.json > packageManager")
		}
	}

	// Yarn
	if p == PackageManagerYarn1 || p == PackageManagerYarn2 {
		if p == PackageManagerYarn1 {
			packages.Default("yarn", "1")
			packages.AddSupportingAptPackage("tar")
			packages.AddSupportingAptPackage("gpg")
		} else {
			packages.Default("yarn", "2")
		}

		name, version := p.parsePackageManagerField(packageJson)
		if name == "yarn" && version != "" {
			majorVersion := strings.Split(version, ".")[0]

			// Only apply version if it matches the expected yarn version
			if (majorVersion == "1" && p == PackageManagerYarn1) ||
				(majorVersion != "1" && p == PackageManagerYarn2) {
				packages.Version(packages.Default("yarn", majorVersion), version, "package.json > packageManager")
			}
		}
	}

	// Bun
	if p == PackageManagerBun {
		bun := packages.Default("bun", "latest")

		name, version := p.parsePackageManagerField(packageJson)
		if name == "bun" && version != "" {
			packages.Version(bun, version, "package.json > packageManager")
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

// parsePackageManagerField parses the packageManager field from package.json
// and returns the name and version as a tuple
func (p PackageManager) parsePackageManagerField(packageJson *PackageJson) (string, string) {
	if packageJson.PackageManager != nil {
		pmString := *packageJson.PackageManager

		// Parse packageManager field which is in format "name@version" or "name@version+sha224.hash"
		parts := strings.Split(pmString, "@")
		if len(parts) == 2 {
			// Split version on '+' to remove SHA hash if present
			versionParts := strings.Split(parts[1], "+")
			return parts[0], versionParts[0]
		}
	}

	return "", ""
}

type YarnRc struct {
	GlobalFolder string `json:"globalFolder"`
}

func (p PackageManager) getYarn2GlobalFolder(ctx *generate.GenerateContext) string {
	var yarnRc YarnRc
	if err := ctx.App.ReadYAML(".yarnrc.yml", &yarnRc); err == nil && yarnRc.GlobalFolder != "" {
		return yarnRc.GlobalFolder
	}

	return "/root/.yarn"
}
