package flutter

import (
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	PubspecPath = "pubspec.yaml"
	FvmrcPath   = ".fvmrc"
	// legacy FVM config location, still present in plenty of repos
	FvmConfigPath = ".fvm/fvm_config.json"
)

type Pubspec struct {
	Name         string         `yaml:"name"`
	Environment  map[string]any `yaml:"environment"`
	Dependencies map[string]any `yaml:"dependencies"`
	// the top-level `flutter:` section holds asset/font config and is only present on Flutter apps
	Flutter map[string]any `yaml:"flutter"`
}

// Reports whether this pubspec describes a Flutter app rather than a plain Dart package.
func (p *Pubspec) UsesFlutterSDK() bool {
	if p.Flutter != nil {
		return true
	}

	// dependencies:
	//   flutter:
	//     sdk: flutter
	sdk, _ := nestedString(p.Dependencies["flutter"], "sdk")
	return sdk == "flutter"
}

// gopkg.in/yaml.v2 decodes a mapping into map[any]any whenever the surrounding value type is
// `any`, so asserting map[string]any alone silently misses every nested lookup.
func nestedString(value any, key string) (string, bool) {
	switch m := value.(type) {
	case map[string]any:
		s, ok := m[key].(string)
		return s, ok
	case map[any]any:
		s, ok := m[key].(string)
		return s, ok
	}

	return "", false
}

// Passed through to the resolver as-is, the same way node engines constraints are.
func (p *Pubspec) FlutterConstraint() string {
	constraint, _ := p.Environment["flutter"].(string)
	return strings.TrimSpace(constraint)
}

func (p *FlutterProvider) readPubspec(ctx *generate.GenerateContext) (*Pubspec, error) {
	if p.pubspec != nil {
		return p.pubspec, nil
	}

	if !ctx.App.HasFile(PubspecPath) {
		return nil, nil
	}

	pubspec := Pubspec{}
	if err := ctx.App.ReadYAML(PubspecPath, &pubspec); err != nil {
		return nil, err
	}

	p.pubspec = &pubspec

	return p.pubspec, nil
}

// FVM pins the SDK for a repo. `.fvmrc` is the current format, `.fvm/fvm_config.json` the legacy one.
// Both are JSON, but the version key was renamed between them.
type fvmConfig struct {
	Flutter       string `json:"flutter"`
	FlutterSdkVer string `json:"flutterSdkVersion"`
}

func getFvmVersion(ctx *generate.GenerateContext) (string, string) {
	for _, path := range []string{FvmrcPath, FvmConfigPath} {
		if !ctx.App.HasFile(path) {
			continue
		}

		config := fvmConfig{}
		if err := ctx.App.ReadJSON(path, &config); err != nil {
			ctx.Logger.LogWarn("Failed to parse %s: %s", path, err.Error())
			continue
		}

		version := strings.TrimSpace(config.Flutter)
		if version == "" {
			version = strings.TrimSpace(config.FlutterSdkVer)
		}

		// FVM allows channel names ("stable", "beta") where we need a version mise can resolve
		if version == "" || isFlutterChannel(version) {
			continue
		}

		return version, path
	}

	return "", ""
}

func isFlutterChannel(version string) bool {
	switch version {
	case "stable", "beta", "master", "main", "dev":
		return true
	}
	return false
}
