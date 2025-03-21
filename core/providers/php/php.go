package php

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	_ "embed"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/node"
	"github.com/stretchr/objx"
)

const (
	DEFAULT_PHP_VERSION  = "8.4"
	DefaultCaddyfilePath = "/Caddyfile"
	COMPOSER_CACHE_DIR   = "/opt/cache/composer"
)

//go:embed Caddyfile
var caddyfileTemplate string

//go:embed start-container.sh
var startContainerScript string

//go:embed php.ini
var phpIniTemplate string

type PhpProvider struct{}

func (p *PhpProvider) Name() string {
	return "php"
}

func (p *PhpProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasMatch("index.php") ||
		ctx.App.HasMatch("composer.json"), nil
}

func (p *PhpProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *PhpProvider) Plan(ctx *generate.GenerateContext) error {
	phpImageStep, err := p.phpImagePackage(ctx)
	if err != nil {
		return err
	}

	configFiles, err := p.getConfigFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config files: %w", err)
	}

	isLaravel := p.usesLaravel(ctx)

	prepare := ctx.NewCommandStep("prepare")
	prepare.AddInput(plan.NewStepLayer(phpImageStep.Name()))
	p.Prepare(ctx, prepare, configFiles)

	extensions := ctx.NewCommandStep("extensions")
	extensions.AddInput(plan.NewStepLayer(prepare.Name()))
	p.InstallExtensions(ctx, extensions)

	composer := ctx.NewCommandStep("install:composer")
	composer.AddInput(plan.NewStepLayer(extensions.Name()))
	p.InstallCompose(ctx, composer)

	// Node (if necessary)
	nodeProvider := node.NodeProvider{}
	isNode, err := nodeProvider.Detect(ctx)
	if err != nil {
		return err
	}

	if isLaravel {
		ctx.Logger.LogInfo("Found Laravel app")
	}

	if isNode {
		err = p.DeployWithNode(ctx, nodeProvider, composer, isLaravel)
		if err != nil {
			return err
		}
	} else {
		// A manual build command will go here
		build := ctx.NewCommandStep("build")
		build.AddInput(plan.NewStepLayer(composer.Name()))
		build.AddCommand(plan.NewCopyCommand("."))
		ctx.Deploy.Base = plan.NewStepLayer(build.Name())

		// ctx.Deploy.Inputs = []plan.Layer{
		// 	plan.NewStepLayer(build.Name()),
		// }
	}

	ctx.Deploy.StartCmd = "/start-container.sh"

	return nil
}

func (p *PhpProvider) Prepare(ctx *generate.GenerateContext, prepare *generate.CommandStepBuilder, configFiles *ConfigFiles) {
	if configFiles.Caddyfile.Filename != "" {
		ctx.Logger.LogInfo("Using custom Caddyfile: %s", configFiles.Caddyfile.Filename)
	}

	if configFiles.PhpIni.Filename != "" {
		ctx.Logger.LogInfo("Using custom php.ini: %s", configFiles.PhpIni.Filename)
	}

	if configFiles.StartContainerScript.Filename != "" {
		ctx.Logger.LogInfo("Using custom start-container.sh: %s", configFiles.StartContainerScript.Filename)
	}

	prepare.Assets["Caddyfile"] = configFiles.Caddyfile.Contents
	prepare.Assets["php.ini"] = configFiles.PhpIni.Contents
	prepare.Assets["start-container.sh"] = configFiles.StartContainerScript.Contents

	prepare.AddEnvVars(map[string]string{
		"APP_ENV":       "production",
		"APP_DEBUG":     "false",
		"APP_LOCALE":    "en",
		"LOG_CHANNEL":   "stderr",
		"LOG_LEVEL":     "debug",
		"SERVER_NAME":   ":80",
		"PHP_INI_DIR":   "/usr/local/etc/php",
		"OCTANE_SERVER": "frankenphp",
		"IS_LARAVEL":    strconv.FormatBool(p.usesLaravel(ctx)),
	})
	prepare.AddCommands([]plan.Command{
		plan.NewExecCommand("mkdir -p /usr/local/etc/php/conf.d"),
		plan.NewExecCommand("mkdir -p /conf.d/"),
		plan.NewFileCommand("/usr/local/etc/php/conf.d/php.ini", "php.ini"),
		plan.NewFileCommand(DefaultCaddyfilePath, "Caddyfile"),
		plan.NewFileCommand("/start-container.sh", "start-container.sh", plan.FileOptions{
			CustomName: "create start container script",
			Mode:       0755,
		}),
	})
	prepare.Secrets = []string{}
}

func (p *PhpProvider) InstallExtensions(ctx *generate.GenerateContext, extensions *generate.CommandStepBuilder) {
	phpExtensions := p.getPhpExtensions(ctx)

	if len(phpExtensions) > 0 {
		extensions.AddCommands([]plan.Command{
			plan.NewExecCommand(fmt.Sprintf("install-php-extensions %s", strings.Join(phpExtensions, " "))),
		})
		extensions.Caches = append(extensions.Caches, ctx.Caches.GetAptCaches()...)
	}
	extensions.Secrets = []string{}
}

func (p *PhpProvider) InstallCompose(ctx *generate.GenerateContext, composer *generate.CommandStepBuilder) {
	composer.Secrets = []string{}
	composer.UseSecretsWithPrefixes([]string{"COMPOSER", "PHP"})
	composer.AddVariables(map[string]string{
		"COMPOSER_FUND":      "0",
		"COMPOSER_CACHE_DIR": COMPOSER_CACHE_DIR,
	})
	if _, err := p.readComposerJson(ctx); err == nil {
		composer.AddCache(ctx.Caches.AddCache("composer", COMPOSER_CACHE_DIR))
		composerFiles := p.ComposerSupportingFiles(ctx)

		// Copy composer from the composer image
		composer.AddCommand(plan.CopyCommand{Image: "composer:latest", Src: "/usr/bin/composer", Dest: "/usr/bin/composer"})

		for _, file := range composerFiles {
			composer.AddCommand(plan.NewCopyCommand(file))
		}

		composer.AddCommands([]plan.Command{
			plan.NewExecCommand("composer install --optimize-autoloader --no-scripts --no-interaction"),
		})
	}
}

func (p *PhpProvider) DeployWithNode(ctx *generate.GenerateContext, nodeProvider node.NodeProvider, composer *generate.CommandStepBuilder, isLaravel bool) error {
	err := nodeProvider.Initialize(ctx)
	if err != nil {
		return err
	}

	ctx.Logger.LogInfo("Installing Node")

	miseStep := ctx.GetMiseStepBuilder()
	nodeProvider.InstallMisePackages(ctx, miseStep)

	install := ctx.NewCommandStep("install:node")
	install.AddInput(plan.NewStepLayer(miseStep.Name()))
	nodeProvider.InstallNodeDeps(ctx, install)

	// prune := ctx.NewCommandStep("prune:node")
	// prune.AddInput(plan.NewStepLayer(install.Name()))
	// nodeProvider.PruneNodeDeps(ctx, prune)

	build := ctx.NewCommandStep("build")
	build.Inputs = []plan.Layer{
		plan.NewStepLayer(composer.Name()),
		plan.NewStepLayer(install.Name(), plan.Filter{
			Include: append([]string{"."}, miseStep.GetOutputPaths()...),
		}),
	}
	nodeProvider.Build(ctx, build)

	if isLaravel {
		build.AddCommands([]plan.Command{
			plan.NewExecShellCommand("mkdir -p storage/framework/{sessions,views,cache,testing} storage/logs bootstrap/cache && chmod -R a+rw storage"),
			plan.NewExecCommand("php artisan config:cache"),
			plan.NewExecCommand("php artisan event:cache"),
			plan.NewExecCommand("php artisan route:cache"),
			plan.NewExecCommand("php artisan view:cache"),
		})
	}

	ctx.Deploy.Base = plan.NewStepLayer(build.Name())

	// ctx.Deploy.AddInputs([]plan.Layer{
	// 	plan.NewStepLayer(build.Name(), plan.Filter{
	// 		Include: []string{"."},
	// 		Exclude: []string{"node_modules", "vendor"},
	// 	}),
	// 	plan.NewStepLayer(prune.Name(), plan.Filter{
	// 		Include: []string{"/app/node_modules"},
	// 	}),
	// })

	return nil
}

func (p *PhpProvider) ComposerSupportingFiles(ctx *generate.GenerateContext) []string {
	patterns := []string{
		"**/composer.json",
		"**/composer.lock",
		"artisan",
	}

	var allFiles []string
	for _, pattern := range patterns {
		files, err := ctx.App.FindFiles(pattern)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, files...)

		dirs, err := ctx.App.FindDirectories(pattern)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, dirs...)
	}

	return allFiles
}

func (p *PhpProvider) getPhpExtensions(ctx *generate.GenerateContext) []string {
	extensions := []string{}

	composerJson, err := p.readComposerJson(ctx)
	if err != nil {
		return extensions
	}

	if require, ok := composerJson["require"].(map[string]interface{}); ok {
		for ext := range require {
			if strings.HasPrefix(ext, "ext-") {
				extensions = append(extensions, strings.TrimPrefix(ext, "ext-"))
			}
		}
	}

	if extensionsVar, _ := ctx.Env.GetConfigVariable("PHP_EXTENSIONS"); extensionsVar != "" {
		extensions = append(extensions, strings.FieldsFunc(extensionsVar, func(r rune) bool {
			return r == ',' || r == ' '
		})...)
	}

	if p.usesLaravel(ctx) {
		// https://laravel.com/docs/12.x/deployment#server-requirements
		extensions = append(extensions,
			"ctype",
			"curl",
			"dom",
			"fileinfo",
			"filter",
			"hash",
			"mbstring",
			"openssl",
			"pcre",
			"pdo",
			"session",
			"tokenizer",
			"xml")
	}

	if dbConnection := ctx.Env.GetVariable("DB_CONNECTION"); dbConnection != "" {
		if dbConnection == "mysql" {
			extensions = append(extensions, "pdo_mysql")
		} else if dbConnection == "pgsql" {
			extensions = append(extensions, "pdo_pgsql")
		}
	}

	if p.needsRedisExtension(ctx, composerJson) {
		extensions = append(extensions, "redis")
	}

	return extensions
}

func (p *PhpProvider) needsRedisExtension(ctx *generate.GenerateContext, composerJson map[string]interface{}) bool {
	// Check if Redis is explicitly mentioned in environment variables
	redisHost := ctx.Env.GetVariable("REDIS_HOST")
	redisUrl := ctx.Env.GetVariable("REDIS_URL")
	cacheDriver := ctx.Env.GetVariable("CACHE_DRIVER")
	sessionDriver := ctx.Env.GetVariable("SESSION_DRIVER")
	queueConnection := ctx.Env.GetVariable("QUEUE_CONNECTION")

	if redisHost != "" || redisUrl != "" ||
		cacheDriver == "redis" || sessionDriver == "redis" || queueConnection == "redis" {
		return true
	}

	// Check for Redis packages in composer.json
	if require, ok := composerJson["require"].(map[string]interface{}); ok {
		for pkg := range require {
			if strings.Contains(pkg, "redis") ||
				strings.Contains(pkg, "predis") {
				return true
			}
		}
	}

	return false
}

func (p *PhpProvider) usesLaravel(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("artisan")
}

type ConfigFiles struct {
	Caddyfile            *generate.TemplateFileResult
	StartContainerScript *generate.TemplateFileResult
	PhpIni               *generate.TemplateFileResult
}

func (p *PhpProvider) getConfigFiles(ctx *generate.GenerateContext) (*ConfigFiles, error) {
	phpRootDir := "/app"
	if variable := ctx.Env.GetVariable("RAILPACK_PHP_ROOT_DIR"); variable != "" {
		phpRootDir = variable
	} else if p.usesLaravel(ctx) {
		phpRootDir = "/app/public"
	}

	data := map[string]interface{}{
		"RAILPACK_PHP_ROOT_DIR": phpRootDir,
		"IS_LARAVEL":            p.usesLaravel(ctx),
	}

	caddyfile, err := ctx.TemplateFiles([]string{"Caddyfile"}, caddyfileTemplate, data)
	if err != nil {
		return nil, err
	}

	startContainerScript, err := ctx.TemplateFiles([]string{"start-container.sh"}, startContainerScript, data)
	if err != nil {
		return nil, err
	}

	phpIni, err := ctx.TemplateFiles([]string{"php.ini"}, phpIniTemplate, data)
	if err != nil {
		return nil, err
	}

	return &ConfigFiles{
		Caddyfile:            caddyfile,
		StartContainerScript: startContainerScript,
		PhpIni:               phpIni,
	}, nil
}

func (p *PhpProvider) phpImagePackage(ctx *generate.GenerateContext) (*generate.ImageStepBuilder, error) {
	imageStep := ctx.NewImageStep("packages:image", func(options *generate.BuildStepOptions) string {
		if phpVersion, ok := options.ResolvedPackages["php"]; ok && phpVersion.ResolvedVersion != nil {
			return getPhpImage(*phpVersion.ResolvedVersion)
		}

		// Return the default if we were not able to resolve the version
		return getPhpImage(DEFAULT_PHP_VERSION)
	})

	imageStep.AptPackages = append(imageStep.AptPackages, "git", "zip", "unzip", "ca-certificates")

	// Include both build and runtime apt packages since we don't have a separate runtime image
	imageStep.AptPackages = append(imageStep.AptPackages, ctx.Config.BuildAptPackages...)
	imageStep.AptPackages = append(imageStep.AptPackages, ctx.Config.Deploy.AptPackages...)

	php := imageStep.Default("php", DEFAULT_PHP_VERSION)

	// Read composer.json to get the PHP version
	if composerJson, err := p.readComposerJson(ctx); err == nil {
		phpVersion := objx.New(composerJson).Get("require.php")
		if phpVersion.IsStr() {
			if strings.HasPrefix(phpVersion.Str(), "^") {
				imageStep.Version(php, strings.TrimPrefix(phpVersion.Str(), "^"), "composer.json > require > php")
			} else {
				imageStep.Version(php, phpVersion.Str(), "composer.json > require > php")
			}
		}
	}

	// Ensure that the version is available on Docker Hub
	imageStep.SetVersionAvailable(php, func(version string) bool {
		image := getPhpImage(version)

		// dunglas/frankenphp:php8.4.3-bookworm -> [dunglas, frankenphp, php8.4.3-bookworm]
		parts := strings.Split(image, ":")
		repository := parts[0] // dunglas/frankenphp
		tag := parts[1]        // php8.4.3-bookworm

		url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/%s", repository, tag)
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	})

	return imageStep, nil
}

func getPhpImage(phpVersion string) string {
	return fmt.Sprintf("dunglas/frankenphp:php%s-bookworm", phpVersion)
}

func (p *PhpProvider) readComposerJson(ctx *generate.GenerateContext) (map[string]interface{}, error) {
	var composerJson map[string]interface{}
	err := ctx.App.ReadJSON("composer.json", &composerJson)
	if err != nil {
		return nil, err
	}

	return composerJson, nil
}

func (p *PhpProvider) StartCommandHelp() string {
	return ""
}
