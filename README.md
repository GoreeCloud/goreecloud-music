# GoreeCloud Music

GoreeCloud Music is GoreeCloud's first-party, self-hosted-first music platform. The project is being developed as original GoreeCloud-owned software with a native application architecture and bounded compatibility layers for established music protocols where appropriate.

## Current implementation state

**Version:** `0.1.0-dev.1`  
**Lifecycle:** Active Development  
**Current milestone:** Milestone 1 — Native Multi-User Library  
**Stable:** No

The current Development source includes the accepted Milestone 0 architecture/storage foundation plus bounded Milestone 1 foundations for SQLite application-state persistence, per-profile library authorization, source-preserving library-file scanning/reconciliation, profile-scoped Favorites/Ratings/Recently Played state, an authorization-scoped Recently Added query, and read-only embedded metadata extraction/persistence for MP3 ID3v2.3/v2.4 and FLAC Vorbis comments.

PR #7 established SQLite persistence, explicit `library_memberships`, profile/library persistence, schema-v1 owner-authorization preservation during migration to schema v2, storage-aware health reporting, and supporting tests. PR #9 advanced the storage schema to version 3 with durable `library_files`, supported-audio discovery, no-symlink traversal, source-preserving Added/Updated/Unchanged/Missing/Restored reconciliation, missing-file tombstones, and permission-gated file-state access. PR #12 added authorization-scoped Favorites, Ratings, and Recently Played persistence. PR #14 added the bounded Recently Added query foundation. PR #16 advanced the schema to version 4 with `library_file_metadata` and added bounded read-only embedded metadata extraction for MP3 ID3v2.3/v2.4 and FLAC Vorbis comments. PR #16 exact candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed Music CI `35030658032` and Platform Contract `35030658657`; it merged as `bc90686a18338afef650b73357f454d1be541afb`, and post-merge `main` passed Music CI `35030857299` and Platform Contract `35030858110`.

This remains a bounded backend foundation, not a complete usable music application. Approved sidecar metadata, artwork extraction/processing, additional embedded-tag/container formats, canonical Recording/Release/Source Item/Playable Asset ingestion, codec/container probing, event-driven filesystem change detection, user-facing Recently Added integration, library/search/profile-state APIs, production authentication/session integration, Everkeep recovery acceptance, playback, provider integration, offline downloads, recommendations, and user-facing clients remain open.

## Development service

```bash
go run ./cmd/musicd
```

Default development bind: `127.0.0.1:8080`.

Implemented development endpoints:

- `GET /healthz`
- `GET /api/v1/system/info`

The service opens and migrates its local Development application-state database at startup. `GOREECLOUD_MUSIC_DB` may override the local database path.

The scanner/reconciliation, embedded-metadata, profile-state, and Recently Added foundations currently exist as internal storage/library capabilities; no user-facing library-import, metadata, profile-state, Recently Added, or scan HTTP endpoint is claimed by this repository state.

See [`api/openapi.yaml`](api/openapi.yaml) for the current API contract.

## Architecture and documentation controls

- [`SPECIFICATIONS.md`](SPECIFICATIONS.md)
- [`FEATURES.md`](FEATURES.md)
- [`FEATURE-ROADMAP.md`](FEATURE-ROADMAP.md)
- [`CHANGELOG.md`](CHANGELOG.md)
- [`docs/architecture/milestone-0.md`](docs/architecture/milestone-0.md)
- [`docs/architecture/milestone-1.md`](docs/architecture/milestone-1.md)
- [`docs/architecture/library-scanning.md`](docs/architecture/library-scanning.md)
- [`docs/architecture/library-metadata.md`](docs/architecture/library-metadata.md)
- [`docs/architecture/storage-model.md`](docs/architecture/storage-model.md)
- [`goreecloud.platform.yaml`](goreecloud.platform.yaml)

The authoritative product-scope and planned-capability record is the Markdown file `GoreeCloud/Projects/Project Specification — Music.md` in the governed GoreeCloud document system. The Drive-side feature roadmap is `GoreeCloud/Feature Roadmap/GoreeCloud Music/FEATURE-ROADMAP.md`; this repository's root `FEATURE-ROADMAP.md` must remain materially synchronized with it. The canonical GoreeCloud product changelog is `GoreeCloud/Changelogs/Change Log — Music.md`, while this repository's `CHANGELOG.md` keeps source-controlled implementation history aligned with verified repository state.

## Storage boundary

The Music storage contract keeps durable application state separate from original user media. The current Development backend is SQLite, but the engine-neutral Music domain/storage contract remains authoritative over backend-local identity. Original user audio files remain independent from application database state, and reusable authentication/provider secrets are not ordinary application records.

The current source verifies forward schema migration through version 4. Schema v2 materializes per-profile library membership while preserving existing schema-v1 library-owner access. Schema v3 adds durable source-file observations used for scanning/reconciliation without storing original media bytes. Schema v4 adds normalized per-file embedded metadata bound to the scanner observation's size/mtime snapshot; it does not establish canonical Recording/Release identity. Favorites, Ratings, Recently Played, and Recently Added use existing application-state entities/records and do not add another schema version. Production persistence qualification, backup/restore acceptance, corruption/recovery procedures, and Everkeep acceptance remain open.

## Provider boundary

No external music provider is enabled. YouTube remains proposed and approval-required; no YouTube adapter, authentication, playback, search, or download capability is implemented by this repository state.

## Security and privacy

Do not commit credentials, provider tokens, private keys, production environment files, private listening data, or user media. See [`SECURITY.md`](SECURITY.md) and [`PRIVACY POLICY.md`](PRIVACY%20POLICY.md).

## Licensing

No open-source license has been authoritatively selected yet. See [`LICENSE`](LICENSE) for the current rights notice.
