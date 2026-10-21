package java

import (
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/internal/utils"
)

const DEFAULT_JDK_VERSION = "21"

func (p *JavaProvider) setJDKVersion(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	jdk := miseStep.Default("java", DEFAULT_JDK_VERSION)
	if jdkVersion, envName := ctx.Env.GetConfigVariable("JDK_VERSION"); jdkVersion != "" {
		miseStep.Version(jdk, jdkVersion, envName)
	}

	if p.usesGradle(ctx) {
		if gradlePkg := miseStep.Resolver.Get("gradle"); gradlePkg != nil {
			// Legacy Gradle releases (e.g. Gradle 5.6.4) cannot run on modern Java 21 and require JDK 8.
			semver, err := utils.ParseSemver(gradlePkg.Version)
			if err == nil && semver.Major <= 5 {
				miseStep.Version(jdk, "8", "Gradle <= 5")
			}
		}
	}
}
