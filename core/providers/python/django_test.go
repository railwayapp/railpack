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
