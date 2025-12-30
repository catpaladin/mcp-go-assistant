# Release Workflow

This project uses automated semantic versioning and release management powered by [semantic-release](https://github.com/semantic-release/semantic-release).

## How It Works

When you merge a pull request to `main`, the release workflow automatically:

1. Analyzes commits using conventional commits
2. Determines the next version (major/minor/patch) based on commit types
3. Updates `CHANGELOG.md` with release notes
4. Creates a Git tag (`vX.Y.Z`)
5. Creates a GitHub release with auto-generated release notes
6. Commits changes back to the repository

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

1. **Checkout**: Fetches full git history
2. **Install semantic-release**: Installs necessary packages
3. **Run semantic-release**: Analyzes commits and creates release
4. **Summary**: Comments on the commit with release information

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

- `.releaserc.json` - semantic-release configuration
- `.commitlintrc.js` - Commitlint configuration for linting commit messages
- `package.json` - Node.js dependencies for release tooling

### Customization

To customize the release behavior:

1. Edit `.releaserc.json` for release rules
2. Edit `.commitlintrc.js` for commit message rules
3. Update scopes in `.commitlintrc.js` to match your project structure

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
3. **Test locally**: Use commitlint to validate commit messages
4. **Monitor releases**: Check GitHub releases for unexpected versions
5. **Keep changelog up to date**: The auto-generated changelog is the source of truth

## Resources

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Release](https://semantic-release.gitbook.io/semantic-release/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
