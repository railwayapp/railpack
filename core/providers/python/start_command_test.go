package python

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestFastAPIStartCommandUsesAvailableServerAndDetectedModule(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []string
		usesUvicorn  bool
	}{
		{
			name:         "explicit uvicorn",
			dependencies: []string{"fastapi>=0.115.5", "uvicorn[standard]>=0.32.1"},
			usesUvicorn:  true,
		},
		{
			name:         "standard extra supplies uvicorn",
			dependencies: []string{"fastapi[standard]==0.115.12"},
			usesUvicorn:  true,
		},
		{
			name:         "standard is the first of multiple extras",
			dependencies: []string{"fastapi[standard,other]>=0.115.12"},
			usesUvicorn:  true,
		},
		{
			name:         "standard is the last of multiple extras",
			dependencies: []string{"fastapi[other,standard]>=0.115.12"},
			usesUvicorn:  true,
		},
		{
			name:         "extras allow whitespace and normalized casing",
			dependencies: []string{"FastAPI [ other , STANDARD ] >=0.115.12"},
			usesUvicorn:  true,
		},
		{
			name:         "a similarly named extra does not supply uvicorn",
			dependencies: []string{"fastapi[standardized]>=0.115.12"},
		},
		{
			name:         "an unrelated package does not identify a FastAPI app",
			dependencies: []string{"fastapi-utils[standard]>=0.2", "uvicorn>=0.32.1"},
		},
		{
			name:         "no server dependency",
			dependencies: []string{"fastapi>=0.115.5"},
		},
	}

	for _, manifestName := range []string{"requirements.txt", "pyproject.toml"} {
		for _, tt := range tests {
			for _, entryFileName := range []string{"main.py", "app.py", "server.py"} {
				t.Run(fmt.Sprintf("%s/%s/%s", manifestName, tt.name, entryFileName), func(t *testing.T) {
					appDirectory := t.TempDir()
					manifestContents := strings.Join(tt.dependencies, "\n") + "\n"
					if manifestName == "pyproject.toml" {
						encodedDependencies, err := json.Marshal(tt.dependencies)
						require.NoError(t, err)
						manifestContents = "[project]\ndependencies = " + string(encodedDependencies) + "\n"
					}
					require.NoError(t, os.WriteFile(filepath.Join(appDirectory, manifestName), []byte(manifestContents), 0644))
					require.NoError(t, os.WriteFile(filepath.Join(appDirectory, entryFileName), []byte("from fastapi import FastAPI\napp = FastAPI()\n"), 0644))
					ctx := testingUtils.CreateGenerateContext(t, appDirectory)
					provider := PythonProvider{}

					expectedStartCommand := "python " + entryFileName
					if tt.usesUvicorn {
						// importing an asgi app does not start a server; uvicorn must load the detected module
						expectedStartCommand = fmt.Sprintf("uvicorn %s:app --host 0.0.0.0 --port ${PORT:-8000}", strings.TrimSuffix(entryFileName, ".py"))
					}
					require.Equal(t, expectedStartCommand, provider.GetStartCommand(ctx))
				})
			}
		}
	}
}

func TestFastAPIStartCommandUsesStructuredProductionDependencies(t *testing.T) {
	tests := []struct {
		name         string
		projectFiles map[string]string
		usesUvicorn  bool
	}{
		{
			name: "Poetry extras array",
			projectFiles: map[string]string{
				"pyproject.toml": "[tool.poetry.dependencies]\nfastapi = {version = \"*\", extras = [\"other\", \"standard\"]}\n",
			},
			usesUvicorn: true,
		},
		{
			name: "Pipenv extras array",
			projectFiles: map[string]string{
				"Pipfile": "[packages]\nfastapi = {version = \"*\", extras = [\"standard\", \"other\"]}\n",
			},
			usesUvicorn: true,
		},
		{
			name: "requirements comments do not install a server",
			projectFiles: map[string]string{
				"requirements.txt": "fastapi>=0.115.5\n# uvicorn\n# fastapi[standard]\n",
			},
		},
		{
			name: "development dependencies do not install a production server",
			projectFiles: map[string]string{
				"pyproject.toml": "[project]\ndependencies = [\"fastapi\"]\n[dependency-groups]\ndev = [\"fastapi[standard]\", \"uvicorn\"]\n",
			},
		},
		{
			name: "unselected optional dependencies do not install a server",
			projectFiles: map[string]string{
				"pyproject.toml": "[project]\ndependencies = [\"fastapi\"]\n[project.optional-dependencies]\nserver = [\"uvicorn\"]\n",
			},
		},
		{
			name: "Poetry development dependencies do not install a production server",
			projectFiles: map[string]string{
				"pyproject.toml": "[tool.poetry.dependencies]\nfastapi = \"*\"\n[tool.poetry.group.dev.dependencies]\nfastapi = {version = \"*\", extras = [\"standard\"]}\nuvicorn = \"*\"\n",
			},
		},
		{
			name: "Pipenv development dependencies do not install a production server",
			projectFiles: map[string]string{
				"Pipfile": "[packages]\nfastapi = \"*\"\n[dev-packages]\nfastapi = {version = \"*\", extras = [\"standard\"]}\nuvicorn = \"*\"\n",
			},
		},
		{
			name: "requirements take precedence over pyproject",
			projectFiles: map[string]string{
				"requirements.txt": "fastapi>=0.115.5\n",
				"pyproject.toml":   "[project]\ndependencies = [\"fastapi[standard]\", \"uvicorn\"]\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appDirectory := t.TempDir()
			for relativePath, fileContents := range tt.projectFiles {
				require.NoError(t, os.WriteFile(filepath.Join(appDirectory, relativePath), []byte(fileContents), 0644))
			}
			require.NoError(t, os.WriteFile(filepath.Join(appDirectory, "server.py"), nil, 0644))
			ctx := testingUtils.CreateGenerateContext(t, appDirectory)
			provider := PythonProvider{}

			expectedStartCommand := "python server.py"
			if tt.usesUvicorn {
				expectedStartCommand = "uvicorn server:app --host 0.0.0.0 --port ${PORT:-8000}"
			}
			require.Equal(t, expectedStartCommand, provider.GetStartCommand(ctx))
		})
	}
}

func TestFrameworkStartCommandsUseDetectedModule(t *testing.T) {
	tests := []struct {
		name                 string
		requirementsContents string
		entryFileName        string
		expectedStartCommand string
	}{
		{
			name:                 "Flask app.py",
			requirementsContents: "flask\ngunicorn\n",
			entryFileName:        "app.py",
			expectedStartCommand: "gunicorn --bind 0.0.0.0:${PORT:-8000} app:app",
		},
		{
			name:                 "Flask server.py",
			requirementsContents: "flask\ngunicorn\n",
			entryFileName:        "server.py",
			expectedStartCommand: "gunicorn --bind 0.0.0.0:${PORT:-8000} server:app",
		},
		{
			name:                 "FastHTML app.py",
			requirementsContents: "python-fasthtml\nuvicorn\n",
			entryFileName:        "app.py",
			expectedStartCommand: "uvicorn app:app --host 0.0.0.0 --port ${PORT:-8000}",
		},
		{
			name:                 "FastHTML server.py",
			requirementsContents: "python-fasthtml\nuvicorn\n",
			entryFileName:        "server.py",
			expectedStartCommand: "uvicorn server:app --host 0.0.0.0 --port ${PORT:-8000}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appDirectory := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(appDirectory, "requirements.txt"), []byte(tt.requirementsContents), 0644))
			require.NoError(t, os.WriteFile(filepath.Join(appDirectory, tt.entryFileName), nil, 0644))
			ctx := testingUtils.CreateGenerateContext(t, appDirectory)
			provider := PythonProvider{}

			require.Equal(t, tt.expectedStartCommand, provider.GetStartCommand(ctx))
		})
	}
}
