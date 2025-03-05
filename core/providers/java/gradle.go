package java

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DEFAULT_GRADLE_VERSION = 8
	GRADLE_CACHE_KEY       = "gradle"
)

func (p *JavaProvider) usesGradle(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("gradlew")
}

func (p *JavaProvider) getGradleVersion(ctx *generate.GenerateContext) (int, error) {
	if envVersion, _ := ctx.Env.GetConfigVariable("GRADLE_VERSION"); envVersion != "" {
		intVersion, err := strconv.Atoi(envVersion)
		if err != nil {
			return DEFAULT_GRADLE_VERSION, err
		} else {
			return intVersion, nil
		}
	}

	if !ctx.App.HasMatch("gradle/wrapper/gradle-wrapper.properties") {
		return DEFAULT_GRADLE_VERSION, nil
	}

	wrapperProps, err := ctx.App.ReadFile("gradle/wrapper/gradle-wrapper.properties")
	if err != nil {
		return DEFAULT_GRADLE_VERSION, err
	}

	versionRegex, err := regexp.Compile(`(distributionUrl[\S].*[gradle])(-)([0-9|\.]*)`)
	if err != nil {
		return DEFAULT_GRADLE_VERSION, err
	}

	if !versionRegex.Match([]byte(wrapperProps)) {
		return DEFAULT_GRADLE_VERSION, nil
	}

	customVersion := string(versionRegex.FindSubmatch([]byte(wrapperProps))[3])

	parseVersionRegex, err := regexp.Compile(`^(?:[\sa-zA-Z-"']*)(\d*)(?:\.*)(\d*)(?:\.*\d*)(?:["']?)$`)
	if err != nil {
		return DEFAULT_GRADLE_VERSION, err
	}

	if !parseVersionRegex.Match([]byte(customVersion)) {
		return DEFAULT_GRADLE_VERSION, nil
	}

	parsedVersion := string(parseVersionRegex.FindSubmatch([]byte(customVersion))[1])

	intVersion, err := strconv.Atoi(parsedVersion)
	if err != nil {
		return DEFAULT_GRADLE_VERSION, err
	} else {
		return intVersion, nil
	}
}

func (p *JavaProvider) installGradle(ctx *generate.GenerateContext) {
	version, _ := p.getGradleVersion(ctx)

	ctx.GetMiseStepBuilder().Default("gradle", strconv.Itoa(version))
}

func (p *JavaProvider) gradleCache(ctx *generate.GenerateContext) string {
	return ctx.Caches.AddCache(GRADLE_CACHE_KEY, "/root/.gradle")
}

func (p *JavaProvider) readBuildGradle(ctx *generate.GenerateContext) string {
	filePath := "build.gradle"
	if !ctx.App.HasMatch(filePath) {
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
