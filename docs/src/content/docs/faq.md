---
title: FAQ
description: Frequently Asked Questions about Railpack
---

## General

### Can Railpack output a Dockerfile?

No. Railpack uses low-level Docker instructions (LLB) rather than generating a Dockerfile.

The philosophy behind Railpack is to abstract away the need to manage Dockerfiles entirely, so we do not plan on supporting the generation or use of Dockerfiles in combination with Railpack.
