---
title: Java
description: Building Java applications with Railpack
---

Railpack builds and deploys Java (including Spring Boot) applications built with Gradle or Maven.

## Detection

Your project will be detected as a Java application if any of these conditions are
met:

- A `gradlew` (Gradle wrapper) file exists in the root directory (to create this, if you don't have one, run `gradle wrapper`)
- A `pom.{xml,atom,clj,groovy,rb,scala,yaml,yml}` file exists in the root directory

## Versions

The Java version is determined in the following order:

- Set via the `RAILPACK_JDK_VERSION` environment variable
- If the project uses Gradle <= 5, Java 8 is used
- Defaults to `21`

### Config Variables

| Variable                  | Description                 | Example |
| ------------------------- | --------------------------- | ------- |
| `RAILPACK_JDK_VERSION`    | Override the JDK version    | `17`    |
| `RAILPACK_GRADLE_VERSION` | Override the Gradle version | `8.5`   |

## Multi-Module Projects

Railpack detects multi-module Maven and Gradle projects (identified by
`<modules>` in the root `pom.xml` or multiple subprojects in Gradle).

For multi-module projects, Railpack finds the JAR to run using this
heuristic:

1. First, looks for `*-jar-with-dependencies.jar` (Maven Assembly Plugin)
2. Then, looks for `*-spring-boot*.jar` (Spring Boot Maven Plugin)
3. Finally, picks the first JAR matching `*/target/*.jar` for Maven or
   `*/build/libs/*.jar` for Gradle

This heuristic may not work if:

- Your project has multiple executable JARs
- JARs don't follow standard naming conventions
- The build produces only thin JARs (without bundled dependencies)

If JAR selection fails, specify a custom start command.

## Custom Start Command

To override the default start command, create a `railpack.json` file in
your project root:

```json
{
  "deploy": {
    "startCommand": "java -jar server/target/my-app.jar"
  }
}
```
