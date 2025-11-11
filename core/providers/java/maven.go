package java

import (
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const MAVEN_CACHE_KEY = "maven"

// isMultimoduleMaven detects if this is a multi-module Maven project where JARs
// are built in subdirectories (e.g., app/target/*.jar) rather than at the root.
func (p *JavaProvider) isMultimoduleMaven(ctx *generate.GenerateContext) bool {
	pomFile, err := ctx.App.ReadFile("pom.xml")
	if err != nil {
		return false
	}
	return strings.Contains(pomFile, "<modules>")
}

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
		// If using the Swarm web server, set the port accordingly for any passed-in $PORT variable
		return "-Dswarm.http.port=$PORT"
	} else if strings.Contains(pomFile, "<groupId>org.springframework.boot") &&
		strings.Contains(pomFile, "<artifactId>spring-boot") {
		// If using Spring Boot, set the port accordingly for any passed-in $PORT variable
		return "-Dserver.port=$PORT"
	}
	return ""
}
