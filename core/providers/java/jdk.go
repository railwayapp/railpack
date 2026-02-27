package java

import (
	"strconv"

	"github.com/railwayapp/railpack/core/generate"
)

const DEFAULT_JDK_VERSION = "21"

func (p *JavaProvider) setJDKVersion(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	jdk := miseStep.Default("java", DEFAULT_JDK_VERSION)
	if jdkVersion, source := p.jdkVersion(ctx); jdkVersion != "" {
		miseStep.Version(jdk, jdkVersion, source)
	}

	if p.usesGradle(ctx) {
		gradleVersion, err := strconv.Atoi(miseStep.Resolver.Get("gradle").Version)
		if err == nil && gradleVersion <= 5 {
			miseStep.Version(jdk, "8", "Gradle <= 5")
		}
	}
}
