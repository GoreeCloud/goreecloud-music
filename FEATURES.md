# GoreeCloud Music — Features and Implementation State

## Implemented in merged source foundations

### Milestone 0 architecture foundation

- Typed domain identities for recordings, releases, source items, playable assets, queue items, profiles, and libraries.
- Source-kind and match-confidence models.
- Derived availability state with independent source identity.
- Provider adapter and registry contracts.
- Authorization request/decision contract.
- Deterministic playback route-selection core with no automatic non-exact substitution by default.
- Stable queue-item model carrying recording/source/route context.
- Engine-neutral application-state schema and forward-only migration contract.
- Minimal Development service with health and build-information endpoints.
- OpenAPI Development contract for implemented endpoints plus separately declared planned API domains.
- Machine-readable GoreeCloud Platform Contract state that does not claim unimplemented integrations.
- Unit tests and GitHub Actions CI definition.

### Milestone 1 persistence and library-authorization foundation

Merged PR #7 establishes the current Development source foundation for:

- SQLite application-state persistence through Go `database/sql` and `modernc.org/sqlite`.
- Logical schema version 2 with explicit `library_memberships`.
- Profile and library persistence.
- Atomic owner membership when a library is created.
- Explicit read, edit, and owner membership evaluation.
- Fail-closed owner downgrade and implicit ownership-grant protections.
- Absolute library-root validation while original media remains outside ordinary database state.
- Schema-v1 to schema-v2 migration that transactionally preserves existing library-owner authorization.
- Persistence across database reopen and explicit default-deny/grant behavior tests.
- Storage-aware `/healthz` behavior.
- Configurable local Development database path through `GOREECLOUD_MUSIC_DB`.

PR #7 merged to `main` as `564ac6a792070996dc39b103232c39b4fca95074`. Post-merge CI `35017128457` and Platform Contract `35017129147` passed on that authoritative revision.

## Partial / foundation only

- Native multi-user library: persistent profile/library membership facts exist, but scanning, media ingestion, metadata/artwork processing, complete multi-user library operations, and library/search APIs do not.
- Playback routing: decision core exists; actual stream acquisition, codec negotiation, transcoding, and sessions do not.
- Provider architecture: contract and registry exist; no real provider adapters exist.
- Authorization: application and durable library-membership foundations exist; production GoreeCloud Identity/Privacy Shield/Wardveil runtime integration does not.
- API: service shell and operational endpoints exist; product library/search/playback APIs are not implemented.
- Queue: domain/schema foundations exist; queue service operations, recovery, and multi-device continuity are not implemented.
- Persistence: the SQLite Development backend is implemented, but production persistence qualification, corruption handling, export/restore, and Everkeep recovery acceptance remain open.

## Planned

All other items remain governed by `FEATURE-ROADMAP.md` and the authoritative project specification. Roadmap presence is not implementation evidence.
