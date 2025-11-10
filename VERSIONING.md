# Versioning Policy

## Current Version Status: v0.x.x

This project is currently in **v0.x.x** release status and should **NOT** be tagged as v1.0.0 or higher until explicitly ready for a stable 1.0 release.

## Version Guidelines

- **v0.x.x**: Active development, breaking changes may occur between minor versions
- **v1.0.0+**: Stable API, semantic versioning strictly followed

## When to Release v1.0.0

The project will be ready for v1.0.0 when:
- [ ] All planned core features are complete (see README.md roadmap)
- [ ] API is stable and well-documented
- [ ] Comprehensive test coverage
- [ ] Production deployment tested on multiple platforms (gokrazy, Docker, standalone)
- [ ] Documentation is complete
- [ ] No known critical bugs

## Current Version: v0.13.0

Latest changes:
- Repository cleanup (removed node_modules and old binary artifacts)
- Fresh git history with proper module path (github.com/drummonds/godocs)
- Pure Go WebAssembly frontend implementation

## For Contributors

When creating new releases:
1. Use `git tag -a v0.x.x -m "version message"` for new versions
2. Increment minor version (0.x) for new features
3. Increment patch version (0.x.x) for bug fixes
4. **Do NOT create v1.0.0 tags** until the criteria above are met

## For Claude Code Sessions

**Important**: This project uses v0.x.x versioning. Do not suggest or create v1.0.0 tags until explicitly instructed by the project maintainer.
