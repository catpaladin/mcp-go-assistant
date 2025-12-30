# Release Workflow

This project uses semi-automated semantic versioning and release management powered by [GoReleaser](https://goreleaser.com/) and conventional commits.

## How It Works

The release process is **semi-automated**:

1. After merging PRs to `main`, run `./scripts/tag-release.sh` locally
2. The script analyzes commits and creates a version tag (vX.Y.Z)
3. Push the tag to trigger the release workflow automatically
4. GoReleaser builds cross-platform binaries
5. GitHub release is created with binaries, checksums, and release notes
6. CHANGELOG.md is automatically updated

## Why This Approach?

This avoids common pitfalls with fully-automated releases:
- ✅ No infinite loops from CHANGELOG commits
- ✅ Full control over when releases happen
- ✅ Easy to test locally before pushing tags
- ✅ Can skip releases by not running the script

## Conventional Commits

The release system relies on conventional commits. Commit messages must follow this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | New feature | Minor |
| `fix` | Bug fix | Patch |
| `perf` | Performance improvement | Patch |
| `docs` | Documentation changes | None |
| `style` | Code style changes (formatting) | None |
| `refactor` | Code refactoring | None |
| `test` | Adding or updating tests | None |
| `build` | Build system or dependency changes | None |
| `ci` | CI configuration changes | None |
| `chore` | Other changes | None |
| `revert` | Revert a previous commit | Patch |

### Breaking Changes

Add `!` after the type/scope to indicate a breaking change:

```
feat!: remove deprecated API endpoint
```

Or add a `BREAKING CHANGE:` footer:

```
feat: add new user authentication

BREAKING CHANGE: changes the authentication flow
```

Breaking changes result in a **major** version bump.

### Examples

```bash
# New feature
git commit -m "feat(codereview): add support for custom guidelines"

# Bug fix
git commit -m "fix: resolve memory leak in rate limiter"

# Breaking change
git commit -m "feat!: remove deprecated configuration options"

# Documentation
git commit -m "docs: update README with installation instructions"

# Test
git commit -m "test: add unit tests for circuit breaker"
```

## Release Workflow

The GitHub Actions workflow (`.github/workflows/release.yml`) is **triggered only on tag pushes**.

### Steps

1. **Local Version Calculation**: Run `./scripts/tag-release.sh` to determine version
2. **Create Tag**: Script creates a tag (vX.Y.Z) and pushes it
3. **Workflow Trigger**: Tag push triggers GitHub Actions
4. **GoReleaser Build**: Builds cross-platform binaries
5. **Create Release**: Publishes GitHub release with binaries and checksums
6. **Update Changelog**: CHANGELOG.md is updated automatically

### GoReleaser Configuration

The `.goreleaser.yml` file configures:
- Multi-platform builds (Linux, macOS, Windows on amd64/arm64)
- Multiple binaries (mcp-go-assistant, client, review-client)
- Automatic checksum generation
- Optional GPG signing (requires `GPG_FINGERPRINT` secret)
- Release notes generated from commit messages

## How to Make a Release

### Quick Start

```bash
# 1. Ensure you're on main and up to date
git checkout main
git pull

# 2. Run the release script (dry run first to see what will happen)
DRY_RUN=true ./scripts/tag-release.sh

# 3. Create the tag and trigger release
./scripts/tag-release.sh
```

### What the Script Does

1. **Analyzes commits** since the last tag
2. **Calculates version** based on commit types:
   - `feat:` → Minor bump (v1.0.0 → v1.1.0)
   - `fix:` or `perf:` → Patch bump (v1.1.0 → v1.1.1)
   - `BREAKING CHANGE:` → Major bump (v1.1.1 → v2.0.0)
3. **Creates tag** with semantic version
4. **Pushes tag** to remote (triggers release workflow)
5. **Updates CHANGELOG.md** with release notes

### Skipping a Release

If you push to `main` but don't want to create a release yet:

```bash
# Just merge normally - release only happens when you run the script
# OR skip CHANGELOG update:
PUSH=false ./scripts/tag-release.sh
```

### Tag Format

Tags follow semantic versioning: `v1.2.3`

Where:
- `1` = Major version (breaking changes)
- `2` = Minor version (new features)
- `3` = Patch version (bug fixes, improvements)

### Release Notes

Release notes are auto-generated from commit messages and include:

- Features added
- Bug fixes
- Performance improvements
- Breaking changes

## Configuration

### Release Configuration

- `.goreleaser.yml` - GoReleaser configuration for builds and releases
- `scripts/tag-release.sh` - Bash script for version calculation
- `.github/workflows/release.yml` - GitHub Actions workflow

### Customization

To customize the release behavior:

1. Edit `.goreleaser.yml` for build configuration:
   - Add/remove target platforms
   - Configure build flags and ldflags
   - Adjust archive formats
   - Enable/disable GPG signing
2. Edit `scripts/tag-release.sh` to modify version calculation rules
3. Update `CHANGELOG.md` template in the workflow if needed

## Troubleshooting

### Script Won't Run

1. Make sure you're on the `main` branch
2. Ensure you have write access to the repository
3. Check that the script is executable: `chmod +x scripts/tag-release.sh`

### Release Not Created

1. Check that commits follow conventional commit format
2. Verify the tag was pushed successfully
3. Check GitHub Actions logs for errors

### Wrong Version Bumped

1. Review commit types since last tag
2. Check for breaking changes (`!` or `BREAKING CHANGE:`)
3. Use `DRY_RUN=true` to preview before creating

### Tag Already Exists

If the tag already exists:
```bash
# Delete local tag
git tag -d vX.Y.Z

# Delete remote tag
git push origin :refs/tags/vX.Y.Z

# Run script again
./scripts/tag-release.sh
```

## Best Practices

1. **Write good commit messages**: Use conventional commits consistently
2. **Review before merging**: Check commit messages in PRs
3. **Test locally**: Use `scripts/tag-release.sh` with `DRY_RUN=true` to preview version
4. **Monitor releases**: Check GitHub releases for unexpected versions
5. **Keep changelog up to date**: The auto-generated changelog is the source of truth

### Local Testing

Test version calculation locally:

```bash
# Dry run to see what version would be created
DRY_RUN=true ./scripts/tag-release.sh

# Actually create a local tag (for testing)
./scripts/tag-release.sh main
```

## Resources

- [Conventional Commits](https://www.conventionalcommits.org/)
- [GoReleaser](https://goreleaser.com/)
- [GoReleaser Documentation](https://goreleaser.com/introduction/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
