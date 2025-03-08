package java

import (
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const MAVEN_CACHE_KEY = "maven"

func (p *JavaProvider) getMavenExe(ctx *generate.GenerateContext) string {
	if ctx.App.HasMatch("mvnw") && ctx.App.HasMatch(".mvn/wrapper/maven-wrapper.properties") {
		return "./mvnw"
	}

	return "mvn"
}

func (p *JavaProvider) mavenCache(ctx *generate.GenerateContext) string {
	return ctx.Caches.AddCache(MAVEN_CACHE_KEY, ".m2/repository")
}

func getMavenPortConfig(ctx *generate.GenerateContext) string {
	pomFile, err := ctx.App.ReadFile("pom.xml")

	if err != nil {
		return ""
	}

	if strings.Contains(pomFile, "<groupId>org.wildfly.swarm") {
		return "-Dswarm.http.port=$PORT"
	} else if strings.Contains(pomFile, "<groupId>org.springframework.boot") &&
		strings.Contains(pomFile, "<artifactId>spring-boot") {
		return "-Dserver.port=$PORT"
	}
	return ""
}
