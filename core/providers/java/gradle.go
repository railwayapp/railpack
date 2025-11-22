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

	versionRegex, err := regexp.Compile(`(distributionUrl[\S].*[gradle])(-)([0-9|\.]*)`)
	if err != nil {
		return
	}

	if !versionRegex.Match([]byte(wrapperProps)) {
		return
	}

	customVersion := string(versionRegex.FindSubmatch([]byte(wrapperProps))[3])

	parseVersionRegex, err := regexp.Compile(`^(?:[\sa-zA-Z-"']*)(\d*)(?:\.*)(\d*)(?:\.*\d*)(?:["']?)$`)
	if err != nil {
		return
	}

	if !parseVersionRegex.Match([]byte(customVersion)) {
		return
	}

	parsedVersion := string(parseVersionRegex.FindSubmatch([]byte(customVersion))[1])

	miseStep.Version(gradle, parsedVersion, "gradle-wrapper.properties")
}

func (p *JavaProvider) gradleCache(ctx *generate.GenerateContext) string {
	return ctx.Caches.AddCache(GRADLE_CACHE_KEY, "/root/.gradle")
}

func (p *JavaProvider) readBuildGradle(ctx *generate.GenerateContext) string {
	filePath := "build.gradle"
	if !ctx.App.HasFile(filePath) {
		filePath = "build.gradle.kts"
	}
	result, err := ctx.App.ReadFile(filePath)
	if err != nil {
		return ""
	} else {
		return result
	}
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
