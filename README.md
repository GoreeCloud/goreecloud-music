# GoreeCloud Music

GoreeCloud Music is GoreeCloud's first-party, self-hosted-first music platform. The project is being developed as original GoreeCloud-owned software with a native application architecture and compatibility boundaries for established music protocols where appropriate.

## Current implementation state

**Version:** `0.1.0-dev.1`  
**Lifecycle:** Active Development  
**Current milestone:** Milestone 1 — Native Multi-User Library  
**Stable:** No

The current `main` source includes the accepted Milestone 0 architecture/storage foundation plus the first bounded Milestone 1 persistence and per-user library-authorization foundation merged through PR #7. It now includes a concrete SQLite Development application-state backend, schema version 2 with explicit `library_memberships`, profile/library persistence, schema-v1 owner-authorization preservation during v2 migration, storage-aware health reporting, tests, exact-source CI, and truthful Platform-System metadata.

This is not a complete usable music library. Incremental scanning, filesystem-change detection, metadata/artwork ingestion, recording/release ingestion, favorites/ratings/recent activity service operations, library/search APIs, production authentication/session integration, Everkeep recovery acceptance, playback, provider integration, offline downloads, recommendations, and user-facing clients remain open.

## Development service

```bash
go run ./cmd/musicd
```

Default development bind: `127.0.0.1:8080`.

Implemented development endpoints:

- `GET /healthz`
- `GET /api/v1/system/info`

The service opens and migrates its local Development application-state database at startup. `GOREECLOUD_MUSIC_DB` may override the local database path.

See [`api/openapi.yaml`](api/openapi.yaml) for the current API contract.

## Architecture controls

- [`SPECIFICATIONS.md`](SPECIFICATIONS.md)
- [`FEATURES.md`](FEATURES.md)
- [`FEATURE-ROADMAP.md`](FEATURE-ROADMAP.md)
- [`docs/architecture/milestone-0.md`](docs/architecture/milestone-0.md)
- [`docs/architecture/milestone-1.md`](docs/architecture/milestone-1.md)
- [`docs/architecture/storage-model.md`](docs/architecture/storage-model.md)
- [`goreecloud.platform.yaml`](goreecloud.platform.yaml)

The authoritative product specification remains the GoreeCloud Drive record `GoreeCloud/Projects/Project Specification — Music.docx`. Repository documentation must remain materially consistent with that record.

## Storage boundary

The Music storage contract keeps durable application state separate from original user media. The current Development backend is SQLite, but the engine-neutral Music domain/storage contract remains authoritative over backend-local identity. Original user audio files remain independent from application database state, and reusable authentication/provider secrets are not ordinary application records.

The current source verifies forward schema migration through version 2, including preservation of existing schema-v1 library-owner authorization. Production persistence qualification, backup/restore acceptance, corruption/recovery procedures, and Everkeep acceptance remain open.

## Provider boundary

No external music provider is enabled. YouTube remains proposed and approval-required; no YouTube adapter, authentication, playback, search, or download capability is implemented by this repository state.

## Security and privacy

Do not commit credentials, provider tokens, private keys, production environment files, private listening data, or user media. See [`SECURITY.md`](SECURITY.md) and [`PRIVACY POLICY.md`](PRIVACY%20POLICY.md).

## Licensing

No open-source license has been authoritatively selected yet. See [`LICENSE`](LICENSE) for the current rights notice.
