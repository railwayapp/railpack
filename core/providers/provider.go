package providers

import (
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/providers/golang"
	"github.com/railwayapp/railpack/core/providers/java"
	"github.com/railwayapp/railpack/core/providers/node"
	"github.com/railwayapp/railpack/core/providers/php"
	"github.com/railwayapp/railpack/core/providers/python"
	"github.com/railwayapp/railpack/core/providers/shell"
	"github.com/railwayapp/railpack/core/providers/staticfile"
)

type Provider interface {
	Name() string
	Detect(ctx *generate.GenerateContext) (bool, error)
	Initialize(ctx *generate.GenerateContext) error
	Plan(ctx *generate.GenerateContext) error
	StartCommandHelp() string
}

func GetLanguageProviders() []Provider {
	return []Provider{
		&php.PhpProvider{},
		&node.NodeProvider{},
		&python.PythonProvider{},
		&golang.GoProvider{},
		&staticfile.StaticfileProvider{},
		&shell.ShellProvider{},
		&java.JavaProvider{},
	}
}

func GetProvider(name string) Provider {
	for _, provider := range GetLanguageProviders() {
		if provider.Name() == name {
			return provider
		}
	}

	return nil
}
