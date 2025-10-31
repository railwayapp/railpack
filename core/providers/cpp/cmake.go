package cpp

import (
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type cmake struct{}

func (_ *CppProvider) DetectCmake(ctx *generate.GenerateContext) (buildSystem, bool) {
	if ctx.App.HasFile("CMakeLists.txt") {
		return &cmake{}, true
	}
	return nil, false
}

func (_ *cmake) Install(ctx *generate.GenerateContext, mise *generate.MiseStepBuilder) {
	mise.Default("cmake", "latest")
	mise.Default("ninja", "latest")
	mise.UseMiseVersions(ctx, []string{"meson", "ninja"})
}

func (_ *cmake) Build(build *generate.CommandStepBuilder) {
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("cmake -B /build -GNinja /app"),
		plan.NewExecCommand("cmake --build /build"),
	})
}
