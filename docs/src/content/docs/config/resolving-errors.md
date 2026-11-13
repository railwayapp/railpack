---
title: Resolving Errors
description: Solutions to common build and deployment errors
---

## Python

### Python version mismatch

If the installed Python version does not satisfy `requires-python` in
`pyproject.toml`, a uv build can fail during `uv sync` with:

```text
error: No interpreter found for Python >=3.11, <3.13 in managed installations or search path

hint: A managed Python download is available for Python >=3.11, <3.13, but Python downloads are set to 'never'
```

`requires-python` declares the versions your project supports. It does not
select the interpreter installed during the build. Without an explicit version,
the default Python version is used, which may fall outside that range.
Builds set `UV_PYTHON_DOWNLOADS=never`, so uv cannot download a replacement.

Choose a version that satisfies the constraint and add a `.python-version`
file at your application's root. For `>=3.11,<3.13`, use Python 3.12:

```text title=".python-version"
3.12
```

Commit the file and rebuild. Alternatively, set the build environment variable
`RAILPACK_PYTHON_VERSION=3.12`. This variable takes precedence over version files,
so update or remove it if it selects an incompatible version.

Keep `requires-python` in `pyproject.toml` to describe your project's supported
versions. See [Python version selection](/languages/python#versions) for all
supported version sources.

## Playwright

### Why does the Playwright installation hang or fail?

Outdated versions of Playwright can cause browser installation to hang, fail,
or behave unexpectedly. Update Playwright and regenerate your lock file with
your package manager so the lock file pins the latest version.

Browser installation is opt-in. Follow the setup instructions for
[Node.js](/languages/node#playwright) or
[Python](/languages/python#playwright) to enable it.
