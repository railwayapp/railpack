package flutter

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "flutter web app",
			path: "../../../examples/flutter-web",
			want: true,
		},
		{
			// detected purely via dependencies > flutter > sdk, with no top-level flutter: section
			name: "flutter web app declaring no assets",
			path: "testdata/no-flutter-section",
			want: true,
		},
		{
			name: "flutter app without a web target",
			path: "testdata/mobile-only",
			want: false,
		},
		{
			name: "dart package that does not use the flutter sdk",
			path: "testdata/dart-package",
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
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := FlutterProvider{}
			got, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPubspec(t *testing.T) {
	tests := []struct {
		name           string
		path           string
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
			name:           "web app with a flutter constraint",
			path:           "testdata/fvm-pinned",
			usesFlutterSDK: true,
			constraint:     ">=3.24.0",
		},
		{
			name:           "plain dart package",
			path:           "testdata/dart-package",
			usesFlutterSDK: false,
			constraint:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
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
		wantVer    string
		wantSource string
	}{
		{
			name:       "fvmrc pin",
			path:       "testdata/fvm-pinned",
			wantVer:    "3.27.1",
			wantSource: FvmrcPath,
		},
		{
			name:       "legacy config pin",
			path:       "testdata/fvm-legacy",
			wantVer:    "3.22.0",
			wantSource: FvmConfigPath,
		},
		{
			// a channel name is not something mise can resolve, so it should be ignored
			name:       "legacy config pinned to a channel",
			path:       "testdata/fvm-channel",
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
			ctx := testingUtils.CreateGenerateContext(t, tt.path)

			version, source := getFvmVersion(ctx)
			require.Equal(t, tt.wantVer, version)
			require.Equal(t, tt.wantSource, source)
		})
	}
}
