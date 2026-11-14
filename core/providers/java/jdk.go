package java

import (
	"strconv"

	"github.com/railwayapp/railpack/core/generate"
)

const DEFAULT_JDK_VERSION = "21"

func (p *JavaProvider) setJDKVersion(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	jdk := miseStep.Default("java", DEFAULT_JDK_VERSION)

	if p.usesGradle(ctx) {
		gradleVersion, err := strconv.Atoi(miseStep.Resolver.Get("gradle").Version)
		if err == nil && gradleVersion <= 5 {
			miseStep.Version(jdk, "8", "Gradle <= 5")
		}
	}
	if p.config.Version != "" {
		miseStep.Version(jdk, p.config.Version, "railpack config > java > version")
	}
}
