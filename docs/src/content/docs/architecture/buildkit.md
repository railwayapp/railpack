---
title: BuildKit LLB Generation
description: Understanding how Railpack generates BuildKit LLB definitions
---

Railpack takes the build plan and generates a BuildKit LLB definition using the
[LLB Go API](https://github.com/moby/buildkit#exploring-llb). The LLB is then
either [sent to the BuildKit daemon](/guides/building-with-cli) or [used by a
custom frontend](/guides/custom-frontend).

Generating LLB directly instead of transpiling the plan into a Dockerfile has
several advantages:

1. **Custom Frontend Integration**: Direct LLB generation enables integration
   with BuildKit's frontend gateway. This allows the platform to either use
   BuildKit through Docker or by interacting with the BuildKit daemon directly.

1. **Caching and Optimization**: Direct LLB generation enables fine-grained
   control over the build cache, allowing more complex caching than what's
   possible with Dockerfile generation.

1. **Secret Management**: LLB provides more secure and flexible secret mounting.

1. **Type Safety and Compile-Time Validation**: The build definition is checked
   at compile-time using the first-party Go library.
