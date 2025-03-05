package java

import (
	"strconv"

	"github.com/railwayapp/railpack/core/generate"
)

const DEFAULT_JDK_VERSION = 21

func (p *JavaProvider) getJDKVersion(ctx *generate.GenerateContext) (int, error) {
	if jdkVersion, _ := ctx.Env.GetConfigVariable("JDK_VERSION"); jdkVersion != "" {
		return strconv.Atoi(jdkVersion)
	}

	if p.usesGradle(ctx) {
		gradleVersion, _ := p.getGradleVersion(ctx)
		if gradleVersion <= 5 {
			return 8, nil
		}
	}

	return DEFAULT_JDK_VERSION, nil
}

func (p *JavaProvider) installJDK(ctx *generate.GenerateContext) error {
	version, err := p.getJDKVersion(ctx)

	if err != nil {
		return err
	}

	ctx.GetMiseStepBuilder().Default("java", strconv.Itoa(version))

	return nil
}
