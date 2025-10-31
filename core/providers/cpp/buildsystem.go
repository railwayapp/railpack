package cpp

import "github.com/railwayapp/railpack/core/generate"

type buildSystem interface {
	Install(pkgs *generate.MiseStepBuilder)
	Build(build *generate.CommandStepBuilder)
}
