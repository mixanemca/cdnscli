# Release Process

This document describes the simple release process for cdnscli.

## Creating a Release

The release process is fully automated via GitHub Actions. To create a new release:

1. **Create and push a git tag:**
   ```bash
   git tag v0.99.0
   git push origin v0.99.0
   ```

2. **GitHub Actions will automatically:**
   - Build binaries for all platforms (Linux, macOS, Windows, amd64, arm64)
   - Create a GitHub Release
   - Upload all binaries as release assets
   - Generate changelog from git commits

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR.MINOR.PATCH** (e.g., `1.0.0`)
- Pre-releases use `-pre` suffix (e.g., `0.99.0`)

## Pre-Release (v0.99.0)

This is a pre-release version to test the release process before v1.0.0.

## Release Checklist

- [ ] All tests pass (`make test`)
- [ ] Version number updated if needed
- [ ] CHANGELOG reviewed (auto-generated from commits)
- [ ] Tag created and pushed
- [ ] GitHub Release created automatically
- [ ] Binaries uploaded to GitHub Releases
- [ ] README installation instructions verified

## Manual Release (if needed)

If you need to create a release manually:

```bash
# Install GoReleaser
brew install goreleaser/tap/goreleaser

# Create a release
goreleaser release --clean
```

