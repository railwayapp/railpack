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

	CaddyfilePath = "/Caddyfile"
	// dart2js resolves imports straight out of PUB_CACHE via absolute paths in
	// .dart_tool/package_config.json, so the package store has to travel with the step layer.
	// A BuildKit cache mount would drop it between the install and build steps.
	PubCacheDir = "/app/.pub-cache"
)

// applied to every step that shells out to flutter. The analytics prompt would otherwise
// appear on first run and the default pub cache lives outside /app.
var flutterVariables = map[string]string{
	"PUB_CACHE":                  PubCacheDir,
	"FLUTTER_SUPPRESS_ANALYTICS": "true",
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
	p.installFlutter(ctx)

	install := p.installStep(ctx)
	build := p.buildStep(ctx, install)

	return p.deploy(ctx, build)
}

func (p *FlutterProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *FlutterProvider) StartCommandHelp() string {
	return "Railpack builds Flutter apps for the web and serves the output with Caddy.\n\n" +
		"The build output is read from build/web. To customize the build, set RAILPACK_BUILD_CMD, e.g.\n" +
		"   RAILPACK_BUILD_CMD=\"flutter build web --wasm --base-href /app/\"\n\n" +
		"To pin the Flutter version, Railpack will check:\n\n" +
		"1. The RAILPACK_FLUTTER_VERSION environment variable\n\n" +
		"2. A mise or asdf version file (mise.toml, .tool-versions)\n\n" +
		"3. An FVM pin (.fvmrc or .fvm/fvm_config.json)\n\n" +
		"4. The \"environment > flutter\" constraint in pubspec.yaml"
}

// installFlutter resolves the Flutter version, from lowest to highest precedence.
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
	install.AddInput(plan.NewLocalLayer())

	install.AddVariables(flutterVariables)

	install.AddCommands([]plan.Command{
		// download the web SDK now so it lands in the layer the build step inherits
		plan.NewExecCommand("flutter precache --web"),
		plan.NewExecCommand("flutter pub get"),
	})

	return install
}

func (p *FlutterProvider) buildStep(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) *generate.CommandStepBuilder {
	build := ctx.NewCommandStep("build")

	// intentionally unfiltered: Flutter precaches dart-sdk and flutter_web_sdk into its mise
	// install dir, and an unfiltered step layer carries /mise forward. Anchoring on the bare
	// mise layer instead would re-download roughly 600 MB on every build.
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
		"DistDir":       fmt.Sprintf("/app/%s", WebOutputDir),
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
		}),
	})

	ctx.Deploy.StartCmd = fmt.Sprintf("caddy run --config %s --adapter caddyfile 2>&1", CaddyfilePath)

	return nil
}
