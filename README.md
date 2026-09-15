# GoreeCloud Music

GoreeCloud Music is GoreeCloud's first-party, self-hosted-first music platform. The project is being developed as original GoreeCloud-owned software with a native application architecture and compatibility boundaries for established music protocols where appropriate.

## Current implementation state

**Version:** `0.1.0-dev.1`  
**Lifecycle:** Active Development  
**Current milestone:** Milestone 0 — Architecture Foundation  
**Stable:** No

The current repository implements only the architecture foundation: core domain identities, availability semantics, provider and authorization contracts, deterministic source routing, queue identity, a minimal health/build-info service, tests, CI, and truthful Platform-System metadata. It does **not** yet implement a usable music library, playback server, provider integration, offline downloads, recommendations, or user-facing clients.

## Development service

```bash
go run ./cmd/musicd
```

Default development bind: `127.0.0.1:8080`.

Implemented development endpoints:

- `GET /healthz`
- `GET /api/v1/system/info`

See [`api/openapi.yaml`](api/openapi.yaml) for the current contract.

## Architecture controls

- [`SPECIFICATIONS.md`](SPECIFICATIONS.md)
- [`FEATURES.md`](FEATURES.md)
- [`FEATURE-ROADMAP.md`](FEATURE-ROADMAP.md)
- [`docs/architecture/milestone-0.md`](docs/architecture/milestone-0.md)
- [`goreecloud.platform.yaml`](goreecloud.platform.yaml)

The authoritative product specification remains the GoreeCloud Drive record `GoreeCloud/Projects/Project Specification — Music.docx`. Repository documentation must remain materially consistent with that record.

## Provider boundary

No external music provider is enabled in Milestone 0. YouTube remains proposed and approval-required; no YouTube adapter, authentication, playback, search, or download capability is implemented by this repository state.

## Security and privacy

Do not commit credentials, provider tokens, private keys, production environment files, private listening data, or user media. See [`SECURITY.md`](SECURITY.md) and [`PRIVACY POLICY.md`](PRIVACY%20POLICY.md).

## Licensing

No open-source license has been authoritatively selected yet. See [`LICENSE`](LICENSE) for the current rights notice.
