package providers

import (
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/cpp"
	"github.com/railwayapp/railpack/core/providers/deno"
	"github.com/railwayapp/railpack/core/providers/dotnet"
	"github.com/railwayapp/railpack/core/providers/elixir"
	"github.com/railwayapp/railpack/core/providers/gleam"
	"github.com/railwayapp/railpack/core/providers/golang"
	"github.com/railwayapp/railpack/core/providers/java"
	"github.com/railwayapp/railpack/core/providers/node"
	"github.com/railwayapp/railpack/core/providers/php"
	"github.com/railwayapp/railpack/core/providers/python"
	"github.com/railwayapp/railpack/core/providers/ruby"
	"github.com/railwayapp/railpack/core/providers/rust"
	"github.com/railwayapp/railpack/core/providers/shell"
	"github.com/railwayapp/railpack/core/providers/staticfile"
)

type Provider interface {
	Name() string
	Detect(ctx *generate.GenerateContext) (bool, error)
	Initialize(ctx *generate.GenerateContext) error
	Plan(ctx *generate.GenerateContext) error
	CleansePlan(buildPlan *plan.BuildPlan)
	StartCommandHelp() string
}

func GetLanguageProviders() []Provider {
	// Order is important here. The first provider that returns true from Detect() will be used.
	return []Provider{
		&php.PhpProvider{},
		&golang.GoProvider{},
		&java.JavaProvider{},
		&rust.RustProvider{},
		&ruby.RubyProvider{},
		&elixir.ElixirProvider{},
		&python.PythonProvider{},
		&deno.DenoProvider{},
		&dotnet.DotnetProvider{},
		&node.NodeProvider{},
		&gleam.GleamProvider{},
		&cpp.CppProvider{},
		&staticfile.StaticfileProvider{},
		&shell.ShellProvider{},
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
