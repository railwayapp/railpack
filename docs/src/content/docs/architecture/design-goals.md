---
title: Design Goals
description: The principles that guide Railpack's architecture and behavior
---

1. **Zero-config by default.** Analyze source code and build with minimal setup.
2. **No Dockerfiles.** Build images directly with BuildKit LLB.
3. **Single builder for all languages.** One tool for Node, Python, Go, PHP, and more.
4. **Backend-agnostic build plans.** Core generates JSON independent of BuildKit.
5. **Built on BuildKit.** Parallel steps, layer cache, and mount caches.
6. **Built on Mise.** Align dev, CI, and production tool versions.
7. **Unopinionated runtime defaults.** No timezone, locale, or app-specific env vars.
8. **Configurable when needed.** CLI flags, env vars, or `railpack.json`.
9. **Stable defaults for platforms.** `--previous` pins versions across builds.
10. **Tested against real projects.** Hundreds of example apps built and run in CI.