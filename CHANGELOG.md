# Changelog

All notable changes to Creuzz will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Creuzz operator
- `OpenCodeAgent` CRD for agent instantiation
- `ClusterAgentTool` CRD for OCI image tools
- `ClusterAgentSkill` CRD for markdown skill injection
- `ClusterAgentMCPServer` CRD for MCP server sidecars
- `ClusterAgentEnv` CRD for environment profiles
- `ClusterAgentEgressRule` CRD for network egress rules
- OCI image volume support (K8s 1.31+)
- Dynamic PATH construction from tool binPaths
- Git bootstrap via init container
- Reactive watches on cluster-scoped resources
- Zero-trust NetworkPolicy generation
- Complete documentation (installation, guides, API reference)

### Changed
- Renamed project from `opencode-operator` to `creuzz`
- Changed API group from `agents.local` to `creuzz.io`

### Fixed
- ConfigMap keys with slashes (use flat keys + volume items)
- Pod recreation on spec changes (delete + recreate)

## [0.1.0] - 2026-10-14

### Added
- Initial project scaffold with Kubebuilder v4
- All 6 CRDs defined and generated
- Complete reconciliation logic
- Unit tests with envtest
- E2E tested on Kind cluster

---

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| Unreleased | - | Documentation, examples, guides |
| 0.1.0 | 2026-10-14 | Initial release |

## Upgrade Guide

### From pre-release to 0.1.0

No upgrade path needed (initial release).

### Future Upgrades

Upgrade instructions will be documented here for each release.
