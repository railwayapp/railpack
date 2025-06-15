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
