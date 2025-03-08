package java

import (
	"strconv"

	"github.com/railwayapp/railpack/core/generate"
)

const DEFAULT_JDK_VERSION = "21"

func (p *JavaProvider) setJDKVersion(ctx *generate.GenerateContext, miseOut *generate.MiseStepBuilder) error {
	miseStep := ctx.GetMiseStepBuilder()
	if miseOut == nil {
		miseOut = miseStep
	}
	jdk := miseOut.Default("java", DEFAULT_JDK_VERSION)
	if jdkVersion, _ := ctx.Env.GetConfigVariable("JDK_VERSION"); jdkVersion != "" {
		miseOut.Version(jdk, jdkVersion, "JDK_VERSION")
	}

	if p.usesGradle(ctx) {
		gradleVersion, err := strconv.Atoi(miseStep.Resolver.Get("gradle").Version)
		if err == nil && gradleVersion <= 5 {
			miseOut.Version(jdk, "8", "Gradle <= 5")
		}
	}
	return nil
}
