package cpp

import (
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type meson struct{}

func (_ *CppProvider) DetectMeson(ctx *generate.GenerateContext) (buildSystem, bool) {
	if ctx.App.HasFile("meson.build") {
		return &meson{}, true
	}
	return nil, false
}

func (_ *meson) Install(mise *generate.MiseStepBuilder) {
	mise.Default("meson", "latest")
	mise.Default("ninja", "latest")
}

func (_ *meson) Build(build *generate.CommandStepBuilder) {
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("meson setup /build"),
		plan.NewExecCommand("meson compile -C /build"),
	})
}
