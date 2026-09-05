---
title: Flutter
description: Deploy Flutter web apps with Railpack
---

Flutter apps are built for the web and served with
[Caddy](https://caddyserver.com/). Only the web target is supported; mobile and
desktop targets are out of scope.

## Detection

Your project is detected as a Flutter web app when all of these are true:

- A `pubspec.yaml` exists in the root directory
- A `web/index.html` exists, meaning the web target is enabled
- `pubspec.yaml` depends on the Flutter SDK, either through
  `dependencies > flutter > sdk: flutter` or a top-level `flutter:` section

If your Flutter app has no `web/` directory, add the web target first:

```bash
flutter create --platforms web .
```

## Version Resolution

The Flutter version is resolved in this order, with later sources taking
precedence:

1. `latest`
2. The `environment > flutter` constraint in `pubspec.yaml`
3. An [FVM](https://fvm.app/) pin in `.fvmrc` or `.fvm/fvm_config.json`
4. A mise or asdf version file (`mise.toml`, `.tool-versions`)
5. The `RAILPACK_FLUTTER_VERSION` environment variable

FVM pins that name a channel (`stable`, `beta`) rather than a version are
ignored, since a channel cannot be resolved to a specific SDK release.

## Build

The build runs in two steps:

```bash
flutter precache --web && flutter pub get
flutter build web --release --no-web-resources-cdn
```

The output is read from `build/web`.

### CanvasKit

`--no-web-resources-cdn` is passed so the app serves CanvasKit from its own
`build/web/canvaskit` directory. Without it, Flutter loads CanvasKit from
`https://www.gstatic.com` at runtime, and a deployed app renders a blank page
with no error whenever that host is unreachable — a corporate proxy, a
restrictive `Content-Security-Policy`, or a region that blocks Google.

The SDK bundles `canvaskit/` into `build/web` either way, so this costs no
extra image size. It does move roughly 2.3 MB (gzipped) per cold page load onto
your own bandwidth. Browsers partition their HTTP cache by site, so the shared
CDN copy would rarely be reused across origins anyway.

To go back to the CDN, override the build:

```bash
RAILPACK_BUILD_CMD="flutter build web --release"
```

### Config Variables

| Variable                   | Description                  | Example  |
| -------------------------- | ---------------------------- | -------- |
| `RAILPACK_FLUTTER_VERSION` | Override the Flutter version | `3.27.1` |

To change the build itself, set `RAILPACK_BUILD_CMD`:

```bash
RAILPACK_BUILD_CMD="flutter build web --wasm --base-href /app/"
```

The output directory is fixed at `build/web`, so a custom build command should
not pass `--output`.

## Serving

Unmatched routes fall back to `index.html` by default, so client-side routes
created with `usePathUrlStrategy()` resolve correctly. Disable this with a
`Staticfile` in your project root:

```yaml
index_fallback: false
```

Flutter's bootstrap files (`index.html`, `flutter_bootstrap.js`,
`flutter_service_worker.js`, `main.dart.js`, `version.json`) are served with
`Cache-Control: no-cache`. These have stable names and are versioned by the
service worker's resource map rather than by a content hash, so caching them
would strand users on a stale build after a deploy.

### Custom Caddyfile

The default
[Caddyfile](https://github.com/railwayapp/railpack/blob/main/core/providers/flutter/Caddyfile.template)
can be replaced with your own `Caddyfile` at the root of your project.

This is how you enable cross-origin isolation, which multi-threaded skwasm
needs when building with `--wasm`:

```caddy
header {
	Cross-Origin-Opener-Policy "same-origin"
	Cross-Origin-Embedder-Policy "require-corp"
}
```

These headers are not set by default because `require-corp` blocks
cross-origin images, fonts, and embeds for apps that do not need it.

## Architecture

Flutter publishes Linux SDK archives for `x64` only, so builds must run on an
`amd64` builder. There is no official `arm64` Linux SDK to install.

## Build Times

The first build downloads the Flutter SDK along with the Dart SDK and web SDK
that `flutter precache` pulls in, roughly 1.5 GB in total. Subsequent builds
reuse the cached layer, but there is no way to avoid the initial download.

Pub packages are installed to `/app/.pub-cache` and travel with the build
layer rather than a build cache mount, because `dart2js` resolves imports from
absolute paths inside the pub cache.

CanvasKit's `*.symbols` files are excluded from the deployed image. They only
symbolicate engine stack traces in a debugger and are never requested at
runtime, so dropping them removes about 8 MB from `build/web`. The renderer
payloads themselves are kept, so a `--wasm` build still has `skwasm` available.

The Flutter SDK is not included in the final image, which contains only Caddy
and the contents of `build/web`.
