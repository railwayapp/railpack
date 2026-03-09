# Release Process

This document outlines the process for creating new releases of Railpack.

## Creating a New Release

1. Determine the new version number following [Semantic
   Versioning](https://semver.org/) principles:

   - MAJOR version for incompatible API changes
   - MINOR version for backwards-compatible functionality additions
   - PATCH version for backwards-compatible bug fixes

2. Create and push a new tag with the version number:

   ```bash
   git fetch origin # or upstream, if you aren't using a fork
   git tag v1.2.3 origin/main  # Replace with your version number
   git push origin v1.2.3
   ```

3. The [release
   workflow](https://github.com/railwayapp/railpack/actions/workflows/release.yml)
   will automatically:
   - Build and publish the frontend Docker image to
     [GHCR](https://github.com/orgs/railwayapp/packages?repo_name=railpack)
   - Create a GitHub release with changelog
   - Build and attach binaries for multiple platforms

### Rerunning a Failed Release

This should *rarely* occur. If you need to rerun a release workflow, on a newer commit:

```
git tag -d v0.18.0 && git push --delete origin v0.18.0 && git tag v0.18.0 origin/main && git push origin v0.18.0
```

## Release Artifacts

### Frontend Docker Image

The frontend Docker image is published to GitHub Container Registry (GHCR) with
the following tags:

- `ghcr.io/railpack/railpack-frontend:latest` (on default branch)
- `ghcr.io/railpack/railpack-frontend:v1.2.3` (specific version)
- `ghcr.io/railpack/railpack-frontend:1.2` (minor version)

The image is built for both `linux/amd64` and `linux/arm64` platforms.

### Binary Releases

The release workflow automatically builds and attaches binaries for multiple
platforms to the GitHub release.

## Verifying a Release

After pushing a tag:

1. Check the [Actions tab](https://github.com/railwayapp/railpack/actions) to
   monitor the release workflow
2. Verify the [GitHub release](https://github.com/railwayapp/railpack/releases)
   is created with the correct artifacts
3. Confirm the frontend Docker image is available in the [package
   registry](https://github.com/railwayapp/railpack/pkgs/container/railpack-frontend)
