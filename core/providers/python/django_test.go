package python

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDjango(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		appName  string
		startCmd string
	}{
		{
			name:     "django project",
			path:     "../../../examples/python-django",
			appName:  "mysite.wsgi",
			startCmd: "python manage.py migrate && gunicorn --bind 0.0.0.0:${PORT:-8000} mysite.wsgi:application",
		},
		{
			name: "non-django project",
			path: "../../../examples/python-uv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PythonProvider{}

			err := provider.Initialize(ctx)
			require.NoError(t, err)

			appName := provider.getDjangoAppName(ctx)
			require.Equal(t, tt.appName, appName)

			startCmd := provider.getDjangoStartCommand(ctx)
			require.Equal(t, tt.startCmd, startCmd)
		})
	}
}

func TestDjangoProviderConfigFromFile(t *testing.T) {
	t.Run("django app name from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/python-django")
		clearConfigVariable(ctx, "DJANGO_APP_NAME")
		setConfigFromJSON(t, ctx, `{
			"python": {
				"djangoAppName": "custom.wsgi"
			}
		}`)

		provider := PythonProvider{}
		require.NoError(t, provider.Initialize(ctx))

		appName := provider.getDjangoAppName(ctx)
		require.Equal(t, "custom.wsgi", appName)

		startCmd := provider.getDjangoStartCommand(ctx)
		require.Equal(t, "python manage.py migrate && gunicorn --bind 0.0.0.0:${PORT:-8000} custom.wsgi:application", startCmd)
	})

	t.Run("django env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/python-django")
		clearConfigVariable(ctx, "DJANGO_APP_NAME")
		ctx.Env.SetVariable("RAILPACK_DJANGO_APP_NAME", "env.wsgi")
		setConfigFromJSON(t, ctx, `{
			"python": {
				"djangoAppName": "custom.wsgi"
			}
		}`)

		provider := PythonProvider{}
		require.NoError(t, provider.Initialize(ctx))

		appName := provider.getDjangoAppName(ctx)
		require.Equal(t, "env.wsgi", appName)

		startCmd := provider.getDjangoStartCommand(ctx)
		require.Equal(t, "python manage.py migrate && gunicorn --bind 0.0.0.0:${PORT:-8000} env.wsgi:application", startCmd)
	})
}
