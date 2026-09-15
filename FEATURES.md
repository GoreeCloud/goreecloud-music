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

### Milestone 1 library scanning and file-reconciliation foundation

Merged PR #9 extends the Development library foundation with:

- Logical schema version 3 and durable `library_files` filesystem observations.
- Forward migration from schema v2 to v3 without changing existing library authorization facts.
- Supported-audio discovery for FLAC, WAV, AIFF/AIF, AAC, M4A, MP3, Opus, Ogg, and Oga extensions.
- No symbolic-link traversal and rejection of a symbolic-link library root.
- Source-preserving scanning that does not write, rename, move, delete, or rewrite media files.
- Library-relative path validation that rejects paths escaping the configured library root.
- Deterministic scan reconciliation for Added, Updated, Unchanged, Missing, and Restored file states.
- Missing-file tombstones through `missing_since` instead of silent identity deletion.
- Read permission for library-file listing and edit permission for scan/reconciliation operations.
- Tests covering discovery, source preservation, authorization isolation, file change/missing/restore behavior, unsafe-path rejection, and schema-v2 to schema-v3 migration.

PR #9 merged to `main` as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`. Its exact candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed CI `35019904248` and Platform Contract `35019905704`; post-merge `main` passed CI `35020139101` and Platform Contract `35020140030`.

### Milestone 1 profile-state foundation

Merged PR #12 adds persistent, profile-owned state for authorized recordings:

- Favorite set/clear operations and bounded favorite listing.
- 0–100 rating set/read/clear operations.
- Recently Played event persistence and bounded recent-history listing.
- Current library-read authorization checks for recording-scoped state mutations and direct rating reads.
- Membership-filtered Favorites and Recently Played queries so revoked library access suppresses stale state from normal Music surfaces.
- Per-profile state isolation when multiple profiles can read the same library.
- Validation for ratings, history timestamps/durations, and result limits.
- Focused tests covering default denial, cross-profile isolation, explicit access grants, authorization revocation, validation, race testing, and service build.

PR #12 merged to `main` as `168d19d07e9b32e0089a512b9e6b7964db76ece1`. Exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed CI `35025974258` and Platform Contract `35025974812`; post-merge `main` passed CI `35026228206` and Platform Contract `35026228918`.

## Partial / foundation only

- Native multi-user library: persistent profile/library membership, source-file observation/reconciliation, and profile-scoped Favorites/Ratings/Recently Played foundations exist, but metadata/artwork extraction, canonical Recording/Release/Source Item/Playable Asset ingestion from scanned files, filesystem event watchers, Recently Added queries, complete multi-user isolation across all remaining library paths, and library/search APIs do not.
- Playback routing: decision core exists; actual stream acquisition, codec negotiation, transcoding, and sessions do not.
- Provider architecture: contract and registry exist; no real provider adapters exist.
- Authorization: application and durable library-membership foundations exist; production GoreeCloud Identity/Privacy Shield/Wardveil runtime integration does not.
- API: service shell and operational endpoints exist; product library/search/playback/profile-state APIs are not implemented.
- Queue: domain/schema foundations exist; queue service operations, recovery, and multi-device continuity are not implemented.
- Persistence: the SQLite Development backend, schema-v3 scan state, and profile-state operations are implemented, but production persistence qualification, corruption handling, export/restore, and Everkeep recovery acceptance remain open.

## Planned

All other items remain governed by `FEATURE-ROADMAP.md` and the authoritative project specification. Roadmap presence is not implementation evidence.
