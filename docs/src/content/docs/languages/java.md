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

The Java version can be set through the `java.version` option in
[`railpack.json`](/config/file#provider-configuration) or the
`RAILPACK_JAVA_VERSION` environment variable. The JSON option takes precedence.
If neither is set, Gradle projects using Gradle <= 5 use Java 8; otherwise the
default is `21`.

### Config Variables

| Variable                          | Description                 | Example |
| --------------------------------- | --------------------------- | ------- |
| `RAILPACK_JAVA_VERSION`           | Override the JDK version    | `17`    |
| `RAILPACK_JAVA_GRADLE_VERSION`    | Override the Gradle version | `8.5`   |

## BuildKit Caching

The Java provider will cache build artifacts &mdash; for Gradle, `~/.gradle` under the key `gradle`, and, for Maven, `.m2/repository` under `maven`.