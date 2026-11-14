// builds a Flutter web target and serves the output with Caddy
// non-web Flutter targets and plain Dart server apps are out of scope

package flutter

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/staticfile"
)

//go:embed Caddyfile.template
var caddyfileTemplate string

const (
	// `flutter build web` always writes here. RAILPACK_BUILD_CMD is the escape hatch for --output
	WebOutputDir = "build/web"
	// the web target is only present when this file exists
	WebIndexPath = "web/index.html"

	CaddyfilePath = "/etc/caddy/Caddyfile"
	// dart2js resolves imports straight out of PUB_CACHE via absolute paths in
	// .dart_tool/package_config.json, so the package store has to travel with the step layer.
	// A BuildKit cache mount would drop it between the install and build steps.
	PubCacheDir = "/app/.pub-cache"
)

var flutterDefaultExcludePatterns = []string{".dart_tool", "build"}

var flutterVariables = map[string]string{
	// Relocates the Dart package store from ~/.pub-cache into /app so packages persist across layers.
	// https://dart.dev/tools/pub/environment-variables
	"PUB_CACHE": PubCacheDir,
	// Disables Flutter's first-run welcome banner and legacy usage telemetry.
	// https://github.com/flutter/flutter/blob/master/packages/flutter_tools/lib/src/runner/flutter_command_runner.dart
	"FLUTTER_SUPPRESS_ANALYTICS": "true",
	// Signals non-interactive bot/CI mode to Flutter's BotDetector and Dart pub (disabling CLI spinners, prompts, and unified analytics).
	// https://github.com/flutter/flutter/blob/master/packages/flutter_tools/lib/src/base/bot_detector.dart
	"CI": "true",
}

type FlutterProvider struct {
	pubspec *Pubspec
}

func (p *FlutterProvider) Name() string {
	return "flutter"
}

func (p *FlutterProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	// a Flutter repo without a web target fails inside the Flutter toolchain with
	// "This application is not configured to build on the web", so let it fall through
	// to the next provider and produce a clearer error instead
	if !ctx.App.HasFile(WebIndexPath) {
		return false, nil
	}

	pubspec, err := p.readPubspec(ctx)
	if err != nil {
		return false, err
	}

	if pubspec == nil {
		return false, nil
	}

	return pubspec.UsesFlutterSDK(), nil
}

func (p *FlutterProvider) Initialize(ctx *generate.GenerateContext) error {
	pubspec, err := p.readPubspec(ctx)
	if err != nil {
		return err
	}

	if pubspec == nil {
		return errors.New("pubspec.yaml could not be found")
	}

	return nil
}

func (p *FlutterProvider) Plan(ctx *generate.GenerateContext) error {
	ctx.SetProviderDefaultExcludes(flutterDefaultExcludePatterns)

	p.installFlutter(ctx)

	install := p.installStep(ctx)
	build := p.buildStep(ctx, install)

	return p.deploy(ctx, build)
}

func (p *FlutterProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *FlutterProvider) StartCommandHelp() string {
	return "Railpack builds Flutter web apps (`flutter build web`) and serves build/web with Caddy.\n" +
		"See the Flutter documentation for configuration options: https://railpack.com/languages/flutter"
}

// Resolves the Flutter version from lowest to highest precedence.
// UseMiseVersions overwrites anything set before it, so the env override is applied last.
func (p *FlutterProvider) installFlutter(ctx *generate.GenerateContext) {
	miseStep := ctx.GetMiseStepBuilder()
	flutter := miseStep.Default("flutter", "latest")

	if p.pubspec != nil {
		if constraint := p.pubspec.FlutterConstraint(); constraint != "" {
			miseStep.Version(flutter, constraint, "pubspec.yaml > environment > flutter")
		}
	}

	if version, source := getFvmVersion(ctx); version != "" {
		miseStep.Version(flutter, version, source)
	}

	miseStep.UseMiseVersions(ctx, []string{"flutter"})

	if envVersion, varName := ctx.Env.GetConfigVariable("FLUTTER_VERSION"); envVersion != "" {
		miseStep.Version(flutter, envVersion, varName)
	}
}

func (p *FlutterProvider) installStep(ctx *generate.GenerateContext) *generate.CommandStepBuilder {
	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepLayer(ctx.GetMiseStepBuilder().Name()))

	install.AddVariables(flutterVariables)

	// Downloading the web SDK before copying manifests ensures code/pubspec changes reuse the precached ~600 MB layer.
	install.AddCommand(plan.NewExecCommand("flutter precache --web"))

	// Copy dependency manifests without the rest of the repo to avoid busting the package layer on source code changes
	for _, file := range p.supportingInstallFiles(ctx) {
		install.AddCommand(plan.NewCopyCommand(file))
	}

	install.AddCommand(plan.NewExecCommand("flutter pub get"))

	return install
}

// Discovers dependency manifests and configuration files needed to install packages.
func (p *FlutterProvider) supportingInstallFiles(ctx *generate.GenerateContext) []string {
	var files []string

	// Project manifest declaring packages, SDK constraints, and plugin registrations
	if ctx.App.HasFile(PubspecPath) {
		files = append(files, PubspecPath)
	}
	// Lockfile pinning exact package versions for deterministic dependency resolution
	if ctx.App.HasFile("pubspec.lock") {
		files = append(files, "pubspec.lock")
	}
	// Localization configuration; pub get invokes synthetic package generation when generate: true
	if ctx.App.HasFile("l10n.yaml") {
		files = append(files, "l10n.yaml")
	}

	return files
}

func (p *FlutterProvider) buildStep(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) *generate.CommandStepBuilder {
	build := ctx.NewCommandStep("build")

	// Carries forward the precached web SDK in /mise and downloaded packages in /app/.pub-cache from install.
	build.AddInput(plan.NewStepLayer(install.Name()))
	build.AddInput(plan.NewLocalLayer())

	build.AddVariables(flutterVariables)

	// Flutter otherwise loads CanvasKit from gstatic.com at runtime, which leaves a deployed app
	// blank when that CDN is unreachable. The SDK already bundles canvaskit/ into build/web, so
	// serving it locally costs no extra image size, and HTTP cache partitioning means the shared
	// CDN cache buys little. RAILPACK_BUILD_CMD can restore the CDN behavior.
	build.AddCommand(plan.NewExecCommand("flutter build web --release --no-web-resources-cdn"))

	return build
}

func (p *FlutterProvider) deploy(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) error {
	// Flutter's usePathUrlStrategy() produces real paths (/settings), which 404 without a fallback
	indexFallback := true
	if configured := staticfile.GetIndexFallback(ctx); configured != nil {
		indexFallback = *configured
	}

	caddyfile, err := ctx.TemplateFiles([]string{"Caddyfile.template", "Caddyfile"}, caddyfileTemplate, map[string]any{
		"DIST_DIR":      fmt.Sprintf("/app/%s", WebOutputDir),
		"IndexFallback": indexFallback,
	})
	if err != nil {
		return err
	}

	if caddyfile.Filename != "" {
		ctx.Logger.LogInfo("Using custom Caddyfile: %s", caddyfile.Filename)
	}

	installCaddyStep := ctx.NewInstallBinStepBuilder("packages:caddy")
	installCaddyStep.Default("caddy", "latest")

	caddy := ctx.NewCommandStep("caddy")
	caddy.AddInput(plan.NewStepLayer(installCaddyStep.Name()))
	caddy.AddCommands([]plan.Command{
		plan.NewFileCommand(CaddyfilePath, "Caddyfile"),
		plan.NewExecCommand(fmt.Sprintf("caddy fmt --overwrite %s", CaddyfilePath)),
	})
	caddy.Assets = map[string]string{
		"Caddyfile": caddyfile.Contents,
	}

	ctx.Logger.LogInfo("Deploying Flutter web build from %s", WebOutputDir)

	ctx.Deploy.AddInputs([]plan.Layer{
		installCaddyStep.GetLayer(),
		plan.NewStepLayer(caddy.Name(), plan.Filter{
			Include: []string{CaddyfilePath},
		}),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{WebOutputDir},
			// CanvasKit ships ~7.6MB of symbol maps used only to symbolicate engine stack traces
			// in a debugger. No renderer requests them at runtime, so they are dead weight in the
			// deployed image. The renderer payloads themselves are kept: a --wasm build via
			// RAILPACK_BUILD_CMD needs skwasm, so only the .symbols files are dropped.
			Exclude: []string{"**/*.symbols"},
		}),
	})

	ctx.Deploy.StartCmd = fmt.Sprintf("caddy run --config %s --adapter caddyfile 2>&1", CaddyfilePath)

	return nil
}
