#!/usr/bin/env bash
#
# Regenerates this example using `flutter create`.
#
# Usage: ./generate.sh
set -euo pipefail

cd "$(dirname "$0")"

command -v flutter >/dev/null 2>&1 || {
  echo "flutter not found in PATH" >&2
  exit 1
}

PROJECT_NAME="railpack_flutter_web"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# Generate into a scratch directory because flutter create refuses to overwrite existing files.
# --platforms web generates only the web runner files (web/index.html, web/manifest.json).
# --empty skips the sample counter app.
(
  cd "$tmp"
  flutter create \
    --project-name "$PROJECT_NAME" \
    --platforms web \
    --empty \
    app
)

# Replace main.dart with the minimal app tested by test.json
cat > "$tmp/app/lib/main.dart" <<'EOF'
import 'package:flutter/material.dart';

void main() {
  runApp(const RailpackApp());
}

class RailpackApp extends StatelessWidget {
  const RailpackApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Railpack Flutter Web',
      home: Scaffold(
        body: Center(
          child: Text(
            'Hello from Flutter Web!',
            style: Theme.of(context).textTheme.headlineMedium,
          ),
        ),
      ),
    );
  }
}
EOF

# Sync into example root, preserving test metadata and documentation
rsync -a --delete \
  --exclude="generate.sh" \
  --exclude="test.json" \
  --exclude="README.md" \
  --exclude=".idea/" \
  --exclude="*.iml" \
  "$tmp/app/" .

echo "regenerated from 'flutter create' ($(flutter --version | head -n 1))"
