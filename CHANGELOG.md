# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-02

### Added

- Initial release
- Bubbletea-based terminal UI
- Status view - daemon health, version, identity
- Capabilities view - browse discovered capabilities
- RPC view - list registered procedures
- Mesh view placeholder (coming soon)
- Logs view placeholder (coming soon)
- REST client for Hecate daemon API
- Keyboard navigation (Tab, 1-5 for views, r to refresh, q to quit)
- Environment variable configuration (`HECATE_URL`)
- Cross-platform builds (Linux, macOS, Windows)
- GitHub Actions CI/CD

### Changed

- Migrated from `macula-io/macula-hecate-tui` to `hecate-social/hecate-tui`
