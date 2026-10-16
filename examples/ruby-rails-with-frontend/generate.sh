#!/usr/bin/env zsh
# Description: autogen this repo using the latest Rails generator with propshaft

set -euo pipefail

root="${0:A:h}"
scratch="$root/tmp/ruby-rails-with-frontend"

rm -rf "$scratch"
mkdir -p "$scratch"

cd "$scratch"

if ! command -v rails >/dev/null 2>&1; then
  echo "Installing latest rails gem..."
  gem install --no-document rails
fi

echo "Generating Rails application..."
rails new app \
  --skip-git \
  --skip-docker \
  --database=sqlite3 \
  --asset-pipeline=propshaft

# Ensure ruby file directive is present in Gemfile to test Mise v2026.9.2+ Gemfile parsing
if ! grep -q "ruby file:" "$scratch/app/Gemfile"; then
  sed -i '' '1s;^;ruby file: ".ruby-version"\n\n;' "$scratch/app/Gemfile"
fi

# Ensure .ruby-version is set to a clean semver
echo "3.4.1" > "$scratch/app/.ruby-version"

# Add a basic home controller with asset inclusion to test Propshaft asset precompilation
cat << 'EOF' > "$scratch/app/app/controllers/home_controller.rb"
class HomeController < ApplicationController
  def index
    render html: "<h1>Hello from Rails #{Rails.version} on Ruby #{RUBY_VERSION}!</h1>".html_safe, layout: true
  end
end
EOF

# Ensure application stylesheet link tag and body style are in place
sed -i '' 's/stylesheet_link_tag :app/stylesheet_link_tag "application"/' "$scratch/app/app/views/layouts/application.html.erb"
cat << 'EOF' >> "$scratch/app/app/assets/stylesheets/application.css"

body {
  background-color: #fafafa;
}
EOF

# Route root to home#index
cat << 'EOF' > "$scratch/app/config/routes.rb"
Rails.application.routes.draw do
  root "home#index"
end
EOF

# Sync generated app into example directory
rsync -a --delete \
  --exclude=generate.sh \
  --exclude=tmp/ \
  --exclude=test.json \
  --exclude=size.json \
  "$scratch/app/" "$root/"

rm -rf "$scratch"
echo "Done generating Rails application in $root"
