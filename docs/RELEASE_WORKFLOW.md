# Release Workflow

This project uses automated semantic versioning and release management powered by [GoReleaser](https://goreleaser.com/) and conventional commits.

## How It Works

When you merge a pull request to `main`, the release workflow automatically:

1. Analyzes commits using conventional commits (via `scripts/tag-release.sh`)
2. Determines the next version (major/minor/patch) based on commit types
3. Runs GoReleaser to build binaries for multiple platforms
4. Creates a GitHub release with auto-generated release notes and binaries
5. Creates a Git tag (`vX.Y.Z`)
6. Updates `CHANGELOG.md` with release notes

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

The release workflow is triggered on pushes to `main` branch.

### Steps

1. **Version Determination**: Analyzes commits using `scripts/tag-release.sh`
2. **GoReleaser Build**: Builds cross-platform binaries
3. **Create Release**: Publishes GitHub release with binaries and checksums
4. **Create Tag**: Tags the commit with semantic version
5. **Update Changelog**: Updates CHANGELOG.md with release notes

### GoReleaser Configuration

The `.goreleaser.yml` file configures:
- Multi-platform builds (Linux, macOS, Windows on amd64/arm64)
- Multiple binaries (mcp-go-assistant, client, review-client)
- Automatic checksum generation
- Optional GPG signing (requires `GPG_FINGERPRINT` secret)
- Release notes generated from commit messages

### Manual Release

To manually trigger a release (not recommended in normal workflow):

```bash
# Ensure you're on main and up to date
git checkout main
git pull

# Make sure all changes are committed
git status

# Push to trigger release
git push
```

### Skip Release

To prevent a release from being created, add `[skip ci]` to the commit message:

```bash
git commit -m "chore: update documentation [skip ci]"
```

## Version Management

### Go Module Version

When a release is created, the Go module version (`go.mod`) is updated automatically.

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

### Release Not Created

1. Check that commits follow conventional commit format
2. Ensure there are no merge commits (merge rebase instead)
3. Verify that the workflow has permission to write to repository

### Wrong Version Bumped

1. Review commit types in the merge
2. Check for breaking changes (`!` or `BREAKING CHANGE:`)
3. Ensure semantic-release configuration is correct

### Git Push Fails After Release

If semantic-release tries to push but fails:

1. Pull latest changes: `git pull --rebase`
2. Resolve conflicts if any
3. Push again

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
