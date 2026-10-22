package java

import (
	"strconv"

	"github.com/railwayapp/railpack/core/generate"
)

const DEFAULT_JDK_VERSION = "21"

func (p *JavaProvider) setJDKVersion(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	jdk := miseStep.Default("java", DEFAULT_JDK_VERSION)
	if jdkVersion, envName := ctx.Env.GetConfigVariable("JDK_VERSION"); jdkVersion != "" {
		miseStep.Version(jdk, jdkVersion, envName)
	}

	if p.usesGradle(ctx) {
		gradleVersion, err := strconv.Atoi(miseStep.Resolver.Get("gradle").Version)
		if err == nil && gradleVersion <= 5 {
			miseStep.Version(jdk, "8", "Gradle <= 5")
		}
	}
	if p.config.Version != "" {
		miseStep.Version(jdk, p.config.Version, "railpack.json > java > version")
	}
}
