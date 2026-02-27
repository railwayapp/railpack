package java

import (
	"fmt"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	javaconfig "github.com/railwayapp/railpack/core/providers/java/config"
)

type JavaProvider struct{}

func (p *JavaProvider) Name() string {
	return "java"
}

func (p *JavaProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasMatch("pom.{xml,atom,clj,groovy,rb,scala,yaml,yml}") || ctx.App.HasMatch("gradlew"), nil
}

func (p *JavaProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *JavaProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *JavaProvider) StartCommandHelp() string {
	return ""
}

func (p *JavaProvider) Plan(ctx *generate.GenerateContext) error {
	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(ctx.GetMiseStepBuilder().Name()))
	build.AddInput(ctx.NewLocalLayer())

	if p.usesGradle(ctx) {
		ctx.Logger.LogInfo("Using Gradle")

		p.setGradleVersion(ctx)
		p.setJDKVersion(ctx, ctx.GetMiseStepBuilder())

		if ctx.App.HasMatch("gradlew") && !ctx.App.IsFileExecutable("gradlew") {
			build.AddCommand(plan.NewExecCommand("chmod +x gradlew"))
		}

		build.AddCommand(plan.NewExecCommand("./gradlew clean build -x check -x test -Pproduction"))
		build.AddCache(p.gradleCache(ctx))
	} else {
		ctx.Logger.LogInfo("Using Maven")

		ctx.GetMiseStepBuilder().Default("maven", "latest")
		p.setJDKVersion(ctx, ctx.GetMiseStepBuilder())

		if ctx.App.HasMatch("mvnw") && !ctx.App.IsFileExecutable("mvnw") {
			build.AddCommand(plan.NewExecCommand("chmod +x mvnw"))
		}

		build.AddCommand(plan.NewExecCommand(fmt.Sprintf("%s -DoutputFile=target/mvn-dependency-list.log -B -DskipTests clean dependency:list install -Pproduction", p.getMavenExe(ctx))))
		build.AddCache(p.mavenCache(ctx))
	}

	runtimeMiseStep := ctx.NewMiseStepBuilder("packages:mise:runtime")
	p.setJDKVersion(ctx, runtimeMiseStep)

	outPath := "target/."
	if ctx.App.HasMatch("**/build/libs/*.jar") || p.usesGradle(ctx) {
		outPath = "."
	}

	ctx.Deploy.AddInputs([]plan.Layer{
		runtimeMiseStep.GetLayer(),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{outPath},
		}),
	})

	ctx.Deploy.StartCmd = p.getStartCmd(ctx)

	p.addMetadata(ctx)

	return nil
}

func (p *JavaProvider) getStartCmd(ctx *generate.GenerateContext) string {
	if p.usesGradle(ctx) {
		buildGradle := p.readBuildGradle(ctx)
		return fmt.Sprintf("java $JAVA_OPTS -jar %s $(ls -1 */build/libs/*jar | grep -v plain)", getGradlePortConfig(buildGradle))
	} else if ctx.App.HasMatch("pom.xml") {
		return fmt.Sprintf("java %s $JAVA_OPTS -jar target/*jar", getMavenPortConfig(ctx))
	} else {
		return "java $JAVA_OPTS -jar target/*jar"
	}

}

func (p *JavaProvider) addMetadata(ctx *generate.GenerateContext) {
	hasGradle := p.usesGradle(ctx)

	if hasGradle {
		ctx.Metadata.Set("javaPackageManager", "gradle")
	} else {
		ctx.Metadata.Set("javaPackageManager", "maven")
	}

	var framework string
	if p.usesSpringBoot(ctx) {
		framework = "spring-boot"
	}

	ctx.Metadata.Set("javaFramework", framework)
}

func (p *JavaProvider) providerConfig(ctx *generate.GenerateContext) *javaconfig.JavaConfig {
	if ctx.Config == nil {
		return nil
	}

	return ctx.Config.Java
}

func (p *JavaProvider) jdkVersion(ctx *generate.GenerateContext) (string, string) {
	if jdkVersion, envName := ctx.Env.GetConfigVariable("JDK_VERSION"); jdkVersion != "" {
		return jdkVersion, envName
	}

	providerConfig := p.providerConfig(ctx)
	if providerConfig != nil && providerConfig.Version != "" {
		return providerConfig.Version, "java.version"
	}

	return "", ""
}

func (p *JavaProvider) gradleVersion(ctx *generate.GenerateContext) (string, string) {
	if gradleVersion, envName := ctx.Env.GetConfigVariable("GRADLE_VERSION"); gradleVersion != "" {
		return gradleVersion, envName
	}

	providerConfig := p.providerConfig(ctx)
	if providerConfig != nil && providerConfig.GradleVersion != "" {
		return providerConfig.GradleVersion, "java.gradleVersion"
	}

	return "", ""
}

func (p *JavaProvider) usesSpringBoot(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("**/spring-boot*.jar") ||
		ctx.App.HasMatch("**/spring-boot*.class") ||
		ctx.App.HasMatch("**/org/springframework/boot/**")
}
