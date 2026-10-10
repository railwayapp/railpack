package java

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DEFAULT_GRADLE_VERSION = "8"
	GRADLE_CACHE_KEY       = "gradle"
)

// Matches the version in a gradle wrapper's distributionUrl, e.g. `9.5.0` in
// `https\://services.gradle.org/distributions/gradle-9.5.0-bin.zip`. The
// version stops at the first non-numeric segment, so a pre-release URL such as
// `gradle-9.0-milestone-1-bin.zip` yields `9.0`: mise has no tag for the
// milestone itself, and the wrapper bootstraps its own distribution anyway.
var gradleDistributionVersionRegex = regexp.MustCompile(`(?m)^\s*distributionUrl\s*=.*gradle-(\d+(?:\.\d+){0,2})[-.]`)

func (p *JavaProvider) usesGradle(ctx *generate.GenerateContext) bool {
	return ctx.App.HasFile("gradlew")
}

func (p *JavaProvider) setGradleVersion(ctx *generate.GenerateContext) {
	miseStep := ctx.GetMiseStepBuilder()
	gradle := miseStep.Default("gradle", DEFAULT_GRADLE_VERSION)

	if envVersion, envName := ctx.Env.GetConfigVariable("GRADLE_VERSION"); envVersion != "" {
		miseStep.Version(gradle, envVersion, envName)
	}

	if !ctx.App.HasFile("gradle/wrapper/gradle-wrapper.properties") {
		return
	}

	wrapperProps, err := ctx.App.ReadFile("gradle/wrapper/gradle-wrapper.properties")
	if err != nil {
		ctx.Logger.LogWarn("Failed to read gradle/wrapper/gradle-wrapper.properties")
		return
	}

	match := gradleDistributionVersionRegex.FindStringSubmatch(wrapperProps)
	if match == nil {
		return
	}

	// The distribution URL pins an exact version, so request that version from
	// mise. Asking for the major alone lets mise pick the newest tag under it,
	// which can be a milestone preview the project never uses.
	miseStep.Version(gradle, match[1], "gradle-wrapper.properties")
}

func (p *JavaProvider) gradleCache(ctx *generate.GenerateContext) string {
	return ctx.Caches.AddCache(GRADLE_CACHE_KEY, "/root/.gradle")
}

func (p *JavaProvider) readBuildGradle(ctx *generate.GenerateContext) string {
	_, result, _ := ctx.App.ReadFirstFileOf("build.gradle", "build.gradle.kts")
	return result
}

func isUsingSpringBoot(buildGradle string) bool {
	return strings.Contains(buildGradle, "org.springframework.boot:spring-boot") ||
		strings.Contains(buildGradle, "spring-boot-gradle-plugin") ||
		strings.Contains(buildGradle, "org.springframework.boot") ||
		strings.Contains(buildGradle, "org.grails:grails-")
}

func getGradlePortConfig(buildGradle string) string {
	if isUsingSpringBoot(buildGradle) {
		return "-Dserver.port=$PORT"
	} else {
		return ""
	}
}
