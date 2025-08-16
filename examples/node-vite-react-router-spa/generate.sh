# Description: autogen this repo from the latest version of the RR template

pnpx create-react-router@latest . --yes

rm Dockerfile
rm -rf .git

# remove the start script otherwise it won't be treated as an SPA
yq eval -i 'del(.scripts.start)' package.json

# disable SSR for SPA build
sed -i '' 's/ssr: true/ssr: false/' react-router.config.ts
