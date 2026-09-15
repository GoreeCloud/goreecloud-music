# Milestone 0 — Architecture Foundation

## Status

Milestone 0 source scope is implemented as an Active Development architecture foundation. This is not a release, production-readiness, or Stable claim. Exact-source validation and normal review/merge controls remain required for every material revision.

## Implemented boundaries

- Typed Recording, Release, Source Item, Playable Asset, Queue Item, and Profile identities.
- Independent source identity and derived availability state.
- Exact/equivalent/alternate/unknown match confidence model.
- Provider-neutral adapter interface and registry.
- Application authorization contract.
- Deterministic route selection that refuses silent non-exact substitution by default.
- Stable queue identity model with owner/source/recording/route fields.
- Minimal native HTTP service exposing health and development build identity.
- API contract that labels implemented paths and lists future API domains separately.
- Engine-neutral durable application-state schema with explicit profile/library ownership boundaries.
- Forward-only schema migration catalog, planning, backend contract, fail-closed validation, and migration tests.
- Machine-readable Platform-System state that truthfully reports integrations as blocked/planned rather than complete.
- Baseline Go tests and CI workflow with exact source-revision verification on pull requests.

## Explicitly not implemented in Milestone 0

A production persistence/database backend, physical database schema, usable library persistence/scanning, user authentication, real provider integrations, media streaming/transcoding, playlists, downloads, recommendations, Music synchronization, clients, Glaze UI surfaces, Platform-System runtime integrations, deployment packaging, and production recovery are not implemented by this milestone.

## Authority boundaries

External source providers remain subordinate adapters. They do not become authorization, privacy, security, recovery, or application-state authorities. Queue/source identities survive provider failure; route selection may change only within the explicit match and authorization policy.

The storage backend is also subordinate to the Music domain model. A future backend may implement the logical storage contract but must not redefine Music identities, user/library ownership, or authorization boundaries.

## Persistence and migrations

The Milestone 0 storage contract is defined in [`storage-model.md`](storage-model.md) and `internal/storage`.

Schema version `1` declares durable application-state entities for schema metadata, profiles, libraries, recordings, releases, provider source mappings, playable-asset metadata, queues/queue items, favorites, ratings, and private play history. Original user media and reusable authentication/provider secrets remain outside ordinary application database state.

Migration `0001-core-application-state` defines the first forward transition from an uninitialized backend to schema version `1`. The catalog and runner validate contiguous forward migration history, reject unsupported/downgrade paths, fail closed on backend errors, and verify the final schema version.

A production persistence engine remains intentionally unselected. Selecting and implementing a concrete backend is a subsequent bounded implementation decision informed by Milestone 1 multi-user library, deployment, backup/restore, portability, integrity, and operational requirements.

## Milestone acceptance boundary

Milestone 0 means the required architecture contracts exist in executable/documented source with applicable validation. It does not mean the planned Music product is usable, released, deployed, production-accepted, or Stable.

Milestone 1 may begin from this foundation, but native library implementation must still supply real persistence, scanning/indexing, authorization, metadata/artwork, favorites/ratings/recent activity, multi-user isolation, and the corresponding tests/evidence before Milestone 1 can be treated as complete.
