#!/bin/sh
echo "build: $(head -1 build-gh-version.txt)"
echo "deploy: $(docker --version)"
# Build sources must not leak into the final image.
if command -v gh >/dev/null; then
  echo "gh leaked into the runtime image"
else
  echo "gh absent at runtime"
fi
