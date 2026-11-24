---
title: Resolving Errors
description: Solutions to common build and deployment errors
---

## Node.js

### Why do pnpm native module builds fail on Node.js 20?

When a dependency such as `better-sqlite3` needs to compile from source, its
installation script invokes `node-gyp`. On Node.js 20, this can fail before
compilation starts with an error from `undici`, such as:

```text
TypeError: webidl.util.markAsUncloneable is not a function
```

or:

```text
TypeError: Cannot read properties of undefined (reading 'bind')
```

The pnpm install step selects `node-gyp` based on the resolved Node.js version,
but edge cases can still produce an incompatible combination. Older Node.js
versions use a compatible `node-gyp` major, such as `node-gyp@12` for Node.js
20.20.2. Compatible newer Node.js versions keep the unversioned install.

If the installed `node-gyp` requires a newer Node.js version than your
application uses, its own dependencies can fail before compilation starts.
For example, `node-gyp` 13 requires Node.js
`^22.22.2 || ^24.15.0 || >=26.0.0` and does not support Node.js 20. Your
application's lock file does not pin this separate global installation.

If your build still shows this mismatch, update your builder or switch to a
supported Node.js LTS release, such as the latest Node.js 24. Set
`RAILPACK_NODE_VERSION=24`, or update your project's
[Node.js version configuration](/languages/node#nodejs). A prebuilt native binary
may bypass compilation, so a successful install on another machine does not
rule out this mismatch.

pnpm's `engineStrict` setting can reject an incompatible package during
installation, but it does not select an older compatible release. Enabling it
alone will not fix the build. See the
[pnpm documentation](https://pnpm.io/settings/cli#enginestrict) and
[issue #727](https://github.com/railwayapp/railpack/issues/727) for details.

## Playwright

### Why does the Playwright installation hang or fail?

Outdated versions of Playwright can cause browser installation to hang, fail,
or behave unexpectedly. Update Playwright and regenerate your lock file with
your package manager so the lock file pins the latest version.

Browser installation is opt-in. Follow the setup instructions for
[Node.js](/languages/node#playwright) or
[Python](/languages/python#playwright) to enable it.
