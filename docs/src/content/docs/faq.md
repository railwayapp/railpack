---
title: FAQ
description: Frequently Asked Questions about Railpack
---

## Can Railpack output a Dockerfile?

No. Railpack uses low-level Docker instructions (LLB) rather than generating a
Dockerfile.

The philosophy behind Railpack is to abstract away the need to manage
Dockerfiles entirely, so we do not plan on supporting the generation or use of
Dockerfiles in combination with Railpack.

If you are looking to customize the build, the best place to do that is a
`railpack.json` file.

## How does Railpack compare to Cloud Native Buildpacks, and others?

If you've used Heroku, Dokku, Cloud Native Buildpacks, and others before, you
may be wondering how they compare:

- Completely open source, unlike the Heroku builder.
- The open source technology is the default builder for Railway, who funds its
  development.
- A single builder for all languages, unlike Cloud Native Buildpacks. Unlike
  CNB, you don't need to determine which builder is well maintained.
- A novel testing strategy. Check out our test suite: every change to the
  builder is tested across hundreds of example projects.
- Built on Docker LLB. This enables many aspects of your build to run in
  parallel.
- Built on mise. This enables you to align your development, CI, and production
  tooling.
