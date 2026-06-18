---
title: Design Goals
description: The principles that guide Railpack's architecture and behavior
---

Design principles that guide our development of Railpack:

1. **Zero config for all popular languages and frameworks.** One tool for Node, Python, Go, PHP, etc and any popular frameworks within those ecosystems. Building a container for a project should require no railpack-specific customization. 
2. **Unopinionated runtime defaults.** No timezone, locale, or other opinionated configuration is included by default. However, all common configuration paths should be easily supported and obvious universal configuration should be included by default.
3. **Make failures obvious.** When a project contains custom configuration that will likely cause a build error, make this obvious to the user.
4. **Allow overrides.** Make it easy for a user to override any default configuration setting.
5. **Integration tests for everything.** Hundreds of example apps built and run in CI. We prefer larger-scoped integration tests, especially when supporting new frameworks, than tightly scoped unit tests.
