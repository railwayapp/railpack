package flutter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func createContextWithFiles(t *testing.T, files map[string]string) *generate.GenerateContext {
	t.Helper()
	dir := t.TempDir()
	for relPath, content := range files {
		fullPath := filepath.Join(dir, relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
	}
	return testingUtils.CreateGenerateContext(t, dir)
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		files map[string]string
		want  bool
	}{
		{
			name: "flutter web app",
			path: "../../../examples/flutter-web",
			want: true,
		},
		{
			// detected purely via dependencies > flutter > sdk, with no top-level flutter: section
			name: "flutter web app declaring no assets",
			files: map[string]string{
				"web/index.html": "",
				"pubspec.yaml": `name: no_flutter_section
dependencies:
  flutter:
    sdk: flutter
  http: ^1.2.0
`,
			},
			want: true,
		},
		{
			name: "flutter app without a web target",
			files: map[string]string{
				"pubspec.yaml": `name: mobile_only
dependencies:
  flutter:
    sdk: flutter
flutter:
  uses-material-design: true
`,
			},
			want: false,
		},
		{
			name: "dart package that does not use the flutter sdk",
			files: map[string]string{
				"web/index.html": "",
				"pubspec.yaml": `name: dart_package
dependencies:
  shelf: ^1.4.0
`,
			},
			want: false,
		},
		{
			name: "node app",
			path: "../../../examples/node-npm",
			want: false,
		},
		{
			name: "static site",
			path: "../../../examples/staticfile-index",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx *generate.GenerateContext
			if tt.path != "" {
				ctx = testingUtils.CreateGenerateContext(t, tt.path)
			} else {
				ctx = createContextWithFiles(t, tt.files)
			}

			provider := FlutterProvider{}
			got, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPubspecSDKAndConstraint(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		files          map[string]string
		usesFlutterSDK bool
		constraint     string
	}{
		{
			name:           "web app without a flutter constraint",
			path:           "../../../examples/flutter-web",
			usesFlutterSDK: true,
			constraint:     "",
		},
		{
			name: "web app with a flutter constraint",
			files: map[string]string{
				"pubspec.yaml": `name: fvm_pinned
environment:
  sdk: ^3.5.0
  flutter: ">=3.24.0"
dependencies:
  flutter:
    sdk: flutter
`,
			},
			usesFlutterSDK: true,
			constraint:     ">=3.24.0",
		},
		{
			name: "plain dart package",
			files: map[string]string{
				"pubspec.yaml": `name: dart_package
environment:
  sdk: ^3.5.0
dependencies:
  shelf: ^1.4.0
`,
			},
			usesFlutterSDK: false,
			constraint:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx *generate.GenerateContext
			if tt.path != "" {
				ctx = testingUtils.CreateGenerateContext(t, tt.path)
			} else {
				ctx = createContextWithFiles(t, tt.files)
			}

			provider := FlutterProvider{}

			pubspec, err := provider.readPubspec(ctx)
			require.NoError(t, err)
			require.NotNil(t, pubspec)

			require.Equal(t, tt.usesFlutterSDK, pubspec.UsesFlutterSDK())
			require.Equal(t, tt.constraint, pubspec.FlutterConstraint())
		})
	}
}

func TestGetFvmVersion(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		files      map[string]string
		wantVer    string
		wantSource string
	}{
		{
			name: "fvmrc pin",
			files: map[string]string{
				".fvmrc": `{"flutter": "3.27.1"}`,
			},
			wantVer:    "3.27.1",
			wantSource: FvmrcPath,
		},
		{
			name: "legacy config pin",
			files: map[string]string{
				".fvm/fvm_config.json": `{"flutterSdkVersion": "3.22.0"}`,
			},
			wantVer:    "3.22.0",
			wantSource: FvmConfigPath,
		},
		{
			// a channel name is not something mise can resolve, so it should be ignored
			name: "legacy config pinned to a channel",
			files: map[string]string{
				".fvm/fvm_config.json": `{"flutterSdkVersion": "stable"}`,
			},
			wantVer:    "",
			wantSource: "",
		},
		{
			name:       "no fvm config",
			path:       "../../../examples/flutter-web",
			wantVer:    "",
			wantSource: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx *generate.GenerateContext
			if tt.path != "" {
				ctx = testingUtils.CreateGenerateContext(t, tt.path)
			} else {
				ctx = createContextWithFiles(t, tt.files)
			}

			version, source := getFvmVersion(ctx)
			require.Equal(t, tt.wantVer, version)
			require.Equal(t, tt.wantSource, source)
		})
	}
}

func TestVersionResolutionPrecedence(t *testing.T) {
	pinnedFiles := map[string]string{
		"web/index.html": "",
		".fvmrc":         `{"flutter": "3.27.1"}`,
		"pubspec.yaml": `name: fvm_pinned
environment:
  sdk: ^3.5.0
  flutter: ">=3.24.0"
dependencies:
  flutter:
    sdk: flutter
`,
	}

	tests := []struct {
		name       string
		path       string
		files      map[string]string
		envs       map[string]string
		wantVer    string
		wantSource string
	}{
		{
			name:       "defaults to latest when unpinned",
			path:       "../../../examples/flutter-web",
			wantVer:    "latest",
			wantSource: "railpack default",
		},
		{
			name:       "fvmrc overrides pubspec environment constraint",
			files:      pinnedFiles,
			wantVer:    "3.27.1",
			wantSource: FvmrcPath,
		},
		{
			name:  "env var overrides fvmrc and pubspec constraint",
			files: pinnedFiles,
			envs: map[string]string{
				"RAILPACK_FLUTTER_VERSION": "3.29.0",
			},
			wantVer:    "3.29.0",
			wantSource: "RAILPACK_FLUTTER_VERSION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx *generate.GenerateContext
			if tt.path != "" {
				ctx = testingUtils.CreateGenerateContext(t, tt.path)
			} else {
				ctx = createContextWithFiles(t, tt.files)
			}
			for k, v := range tt.envs {
				ctx.Env.SetVariable(k, v)
			}

			provider := FlutterProvider{}
			require.NoError(t, provider.Initialize(ctx))
			require.NoError(t, provider.Plan(ctx))

			pkg := ctx.Resolver.Get("flutter")
			require.NotNil(t, pkg)
			require.Equal(t, tt.wantVer, pkg.Version)
			require.Equal(t, tt.wantSource, pkg.Source)
		})
	}
}
