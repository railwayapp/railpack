package providers

import (
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/providers/golang"
	"github.com/railwayapp/railpack/core/providers/node"
	"github.com/railwayapp/railpack/core/providers/php"
	"github.com/railwayapp/railpack/core/providers/python"
	"github.com/railwayapp/railpack/core/providers/ruby"
	"github.com/railwayapp/railpack/core/providers/staticfile"
)

type Provider interface {
	Name() string
	Detect(ctx *generate.GenerateContext) (bool, error)
	Plan(ctx *generate.GenerateContext) error
}

func GetLanguageProviders() []Provider {
	return []Provider{
		&golang.GoProvider{},
		&node.NodeProvider{},
		&php.PhpProvider{},
		&python.PythonProvider{},
		&ruby.RubyProvider{},
		&staticfile.StaticfileProvider{},
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
