package python

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/resolver"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestRequiresPythonSuggestsExplicitVersionWhenUsingDefault(t *testing.T) {
	const pyproject = "[project]\nrequires-python = \">=3.11,<3.13\"\n"
	tests := []struct {
		name                    string
		pyprojectContents       string
		additionalFiles         map[string]string
		configuredPythonVersion string
		expectedVersion         string
		expectedSource          string
		expectSuggestion        bool
	}{
		{
			name:              "constraint with no version selection",
			pyprojectContents: pyproject,
			expectSuggestion:  true,
		},
		{
			// these operators verify the tip does not depend on interpreting version constraints
			name:              "compatible release and exclusion specifiers",
			pyprojectContents: "[project]\nrequires-python = \"~=3.12.0,!=3.12.1\"\n",
			expectSuggestion:  true,
		},
		{
			name: "no pyproject",
		},
		{
			name:              "no requires-python",
			pyprojectContents: "[project]\nname = \"example\"\n",
		},
		{
			name:              "blank requires-python",
			pyprojectContents: "[project]\nrequires-python = \"  \"\n",
		},
		{
			name:              "requires-python outside project metadata",
			pyprojectContents: "[tool.example]\nrequires-python = \">=3.11,<3.13\"\n",
		},
		{
			name:              "invalid pyproject",
			pyprojectContents: "[project",
		},
		{
			name:                    "environment selects interpreter",
			pyprojectContents:       pyproject,
			configuredPythonVersion: "3.12",
			expectedVersion:         "3.12",
			expectedSource:          "RAILPACK_PYTHON_VERSION",
		},
		{
			name:              "idiomatic file selects interpreter",
			pyprojectContents: pyproject,
			additionalFiles:   map[string]string{".python-version": "3.12.13\n"},
			expectedVersion:   "3.12.13",
			expectedSource:    "idiomatic-version-file",
		},
		{
			name:              "runtime file selects interpreter",
			pyprojectContents: pyproject,
			additionalFiles:   map[string]string{"runtime.txt": "python-3.12.13\n"},
			expectedVersion:   "3.12.13",
			expectedSource:    "runtime.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// resolve macos /var -> /private/var so mise's source paths match the test application's path
			appDirectory, err := filepath.EvalSymlinks(t.TempDir())
			require.NoError(t, err)
			if tt.pyprojectContents != "" {
				require.NoError(t, os.WriteFile(filepath.Join(appDirectory, "pyproject.toml"), []byte(tt.pyprojectContents), 0644))
			}
			for relativePath, fileContents := range tt.additionalFiles {
				require.NoError(t, os.WriteFile(filepath.Join(appDirectory, relativePath), []byte(fileContents), 0644))
			}

			ctx := testingUtils.CreateGenerateContext(t, appDirectory)
			if tt.configuredPythonVersion != "" {
				ctx.Env.SetVariable("RAILPACK_PYTHON_VERSION", tt.configuredPythonVersion)
			}
			provider := PythonProvider{}
			provider.InstallMisePackages(ctx, ctx.GetMiseStepBuilder())

			expectedVersion, expectedSource := DEFAULT_PYTHON_VERSION, resolver.DefaultSource
			if tt.expectedVersion != "" {
				expectedVersion, expectedSource = tt.expectedVersion, tt.expectedSource
			}
			pythonPackage := ctx.Resolver.Get("python")
			require.NotNil(t, pythonPackage)
			require.Equal(t, expectedVersion, pythonPackage.Version)
			require.Equal(t, expectedSource, pythonPackage.Source)

			var suggestions []logger.Msg
			for _, message := range ctx.Logger.Logs {
				if message.Level == logger.Suggestion {
					suggestions = append(suggestions, message)
				}
			}
			if !tt.expectSuggestion {
				require.Empty(t, suggestions)
				return
			}
			require.Len(t, suggestions, 1)
			require.Equal(t, "/config/resolving-errors#python-version-mismatch", suggestions[0].DocsPath)
			require.Contains(t, suggestions[0].Msg, "requires-python")
			require.Contains(t, suggestions[0].Msg, ".python-version")
			require.Contains(t, suggestions[0].Msg, "RAILPACK_PYTHON_VERSION")
		})
	}
}
