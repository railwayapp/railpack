# flutter-web

Minimal Flutter web app. `flutter build web` writes to `build/web`, which is
served by Caddy with index fallback enabled so client-side routes resolve.

`pubspec.lock` is intentionally not committed: it pins transitive package
hashes that would go stale against whatever Flutter SDK the build resolves.
