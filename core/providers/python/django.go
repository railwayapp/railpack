package python

import (
	"fmt"
	"regexp"

	"github.com/railwayapp/railpack/core/generate"
)

func (p *PythonProvider) getDjangoAppName(ctx *generate.GenerateContext) string {
	if appName, _ := ctx.Env.GetConfigVariable("DJANGO_APP_NAME"); appName != "" {
		return appName
	}

	paths, err := ctx.App.FindFiles("**/*.py")
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`WSGI_APPLICATION = ["'](.*).application["']`)

	for _, path := range paths {
		contents, err := ctx.App.ReadFile(path)
		if err != nil {
			continue
		}

		matches := re.FindStringSubmatch(contents)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func (p *PythonProvider) getDjangoStartCommand(ctx *generate.GenerateContext) string {
	appName := p.getDjangoAppName(ctx)
	if appName == "" {
		return ""
	}

	ctx.Logger.LogInfo("Using Django app: %s", appName)
	return fmt.Sprintf("python manage.py migrate && gunicorn --bind 0.0.0.0:${PORT:-8000} %s:application", appName)
}

func (p *PythonProvider) isDjango(ctx *generate.GenerateContext) bool {
	hasManage := ctx.App.HasFile("manage.py")
	importsDjango := p.usesDep(ctx, "django")

	return hasManage && importsDjango
}
