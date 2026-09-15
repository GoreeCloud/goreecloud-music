# GoreeCloud Music Storage Model and Migration Contract

## Status

The accepted Milestone 0 foundation remains the engine-neutral durable application-state contract. Milestone 1 PR #7 merged a concrete SQLite Development backend to `main` as `564ac6a792070996dc39b103232c39b4fca95074` while preserving that logical contract and its replaceability boundary. PR #9 advanced the logical schema to version 3 with source-preserving `library_files`, and PR #16 advanced it to version 4 with `library_file_metadata` bound to exact scanner observations.

SQLite is the implemented Development application-state backend in source. It is not production-authoritative merely because it satisfies the logical schema or passes repository validation. Production persistence qualification, backup/restore acceptance, corruption/recovery evidence, and release acceptance remain separate gates.

## Purpose

GoreeCloud Music separates durable application state from original user media. Milestone 0 established the logical state model and forward-only migration contract before a database engine was selected. Milestone 1 may implement a concrete backend only when it preserves the same ownership, privacy, portability, recovery, and migration boundaries.

The engine-neutral executable contract lives in `internal/storage`. The current merged SQLite Development implementation lives in `internal/storage/sqlite`.

## Storage boundary

The application-state backend may persist identities, normalized metadata, ownership and authorization references, queue state, ratings, favorites, history, migration metadata, filesystem observations, and durable media references. It must not become the authoritative storage location for original user-owned audio files.

Original media remains in GoreeCloud-controlled library storage and is referenced by durable identity and path. Losing or replacing the application-state backend must not by itself imply loss of the original media library.

Reusable provider credentials, private keys, authentication secrets, and equivalent sensitive credentials remain outside this storage model. A production implementation must use the applicable secret or identity authority rather than embedding reusable secrets in ordinary application records.

## Current logical schema — version 4

The current engine-neutral schema contains these entities:

| Entity | Ownership boundary | Purpose |
| --- | --- | --- |
| `schema_metadata` | system | schema version and migration metadata |
| `profiles` | profile | Music profile application state without authentication credentials |
| `libraries` | library | library identity, ownership, and source-root metadata |
| `library_memberships` | library | per-profile library authorization and role state |
| `library_files` | library | source-preserving filesystem observations used for incremental reconciliation |
| `library_file_metadata` | library | normalized embedded metadata bound to a scanner file-observation snapshot |
| `recordings` | library | canonical recording identity and library metadata |
| `releases` | library | release and edition identity |
| `source_items` | library | provider-specific source identity mappings |
| `playable_assets` | library | encoded asset metadata, quality, path reference, and availability facts; never original media bytes |
| `queues` | profile | profile-owned durable queue identity and revision state |
| `queue_items` | profile | queue ordering, recording/source intent, selected route, and route reason |
| `favorites` | profile | profile-scoped favorite state |
| `ratings` | profile | profile-scoped rating state |
| `play_history` | profile | profile-scoped listening history subject to privacy and retention controls |

Schema version 2 introduced explicit per-profile library membership state. Schema version 3 added durable source-file observations. Schema version 4 adds durable extracted metadata facts tied to those observations. The logical model remains independent from SQLite. Future backend migration must preserve these product identities and authorization boundaries rather than exposing backend-local identity as the Music domain model.

## Migration contract

The migration catalog is forward-only and versioned independently from the GoreeCloud Music software release number.

The accepted initial migration is `0001-core-application-state`, moving an uninitialized backend from schema version `0` to `1`.

PR #7 added `0002-library-memberships`, moving schema version `1` to `2` and creating the explicit per-profile library authorization relation. The SQLite implementation performs the table creation and owner-membership backfill in the same transaction, preserving existing schema-v1 library owners before the stored schema version advances. A regression test constructs real schema-v1 SQLite state, upgrades it to the current backend, and verifies that owner authorization and library visibility survive the migration.

PR #9 added `0003-library-files`, moving schema version `2` to `3` with source-preserving filesystem observations for incremental scan reconciliation. The migration changes application-state schema only and does not mutate original media.

PR #16 added `0004-library-file-metadata`, moving schema version `3` to `4` with normalized embedded metadata keyed by scanner `file_id`. A regression test constructs real schema-v3 library/file state, opens it through the current backend, verifies migration to the current schema, confirms the existing file observation survives unchanged, and confirms metadata is absent until explicitly extracted.

The catalog validator fails closed when migration identifiers are missing or duplicated, versions are non-contiguous or non-forward, changes are empty or unsupported, a change references an unknown entity, an entity is created more than once, the current schema is not fully represented, or the catalog does not end at the declared current schema version.

Downgrades are not implicitly supported. Rollback must be designed around data safety, backup and restoration, version compatibility, and explicit recovery evidence rather than inferred by reversing migrations.

## Backend contract

A concrete backend implements the minimal `storage.Backend` boundary:

- report its current schema version;
- apply one declared migration atomically;
- update stored schema version only after the migration succeeds.

The shared migration runner verifies the final backend version and fails when the requested target is not reached. A backend must not silently coerce unsupported schema, skip unknown migrations, or mark a failed migration as applied.

## Milestone 1 SQLite Development foundation

The merged Development source uses SQLite as the concrete GoreeCloud Music application-state backend through Go `database/sql` with `modernc.org/sqlite` v1.58.0. The dependency graph is explicitly pinned, including the SQLite module's required matching `modernc.org/libc` v1.75.6 dependency.

The selection is bounded to application-state persistence. It does not make SQLite a permanent GoreeCloud identity or prevent later migration when verified requirements justify another backend.

The implementation configures:

- foreign-key enforcement;
- WAL journaling;
- full synchronous durability;
- SQLite defensive mode;
- a bounded busy timeout;
- transactional schema migrations;
- schema-v1 to schema-v2 owner-authorization preservation;
- schema-v2 to schema-v3 source-file observation state;
- schema-v3 to schema-v4 embedded metadata state;
- STRICT tables for the current physical schema;
- a local Development database path configurable through `GOREECLOUD_MUSIC_DB`.

The implementation also provides profile and library persistence, atomic owner-membership creation, explicit read/edit/owner membership checks, absolute library-root validation, persistent membership state across reopen, source-preserving scanner observations, bounded extracted metadata state, profile-scoped Favorites/Ratings/Recently Played state, authorization-scoped Recently Added retrieval, and storage-aware health reporting.

Owner authority fails closed: an owner cannot be silently downgraded and owner permission cannot be granted to another profile without a separately designed ownership-transfer operation.

For metadata source access, the implementation roots file access with Go `os.Root`, rejects symbolic-link path components under the stricter Music policy, verifies that the final file is regular, and validates scanner size/mtime before parsing and again immediately before metadata persistence. This is a freshness boundary, not a content-integrity hash.

PR #7's exact merge candidate `7113754cd130bbdb6591de757eb51e31dd8541e1` passed Music CI `35016356483` and Platform Contract `35016357625`. After merge, authoritative `main` commit `564ac6a792070996dc39b103232c39b4fca95074` passed post-merge Music CI `35017128457` and Platform Contract `35017129147`.

PR #9 merged the schema-v3 scanner state as authoritative `main` commit `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`, which passed post-merge Music CI `35020139101` and Platform Contract `35020140030`.

PR #16 exact candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed Music CI `35030658032` and Platform Contract `35030658657`. It merged as signed authoritative `main` commit `bc90686a18338afef650b73357f454d1be541afb`, which passed post-merge Music CI `35030857299` and Platform Contract `35030858110`.

## Why SQLite is appropriate for this bounded Development foundation

The current self-hosted Music service is a single application authority rather than a distributed multi-writer database cluster. SQLite provides transactional relational state, explicit schema/migrations, indexes and constraints, a small operational footprint, file-level portability, and simple self-hosted deployment while keeping original media independent from the database.

This choice remains subject to workload, concurrency, recovery, corruption, observability, and production-acceptance evidence. If later Music requirements exceed the verified operating envelope, the engine-neutral storage contract permits a controlled migration to another backend without redefining Music identities or authorization semantics.

## Everkeep and recovery boundary

Selecting, merging, and validating a Development backend is not backup or recovery acceptance. Before release, the SQLite application-state database and required schema/migration metadata must become part of an Everkeep-aligned backup and restore plan with tested restoration evidence.

Original media remains a separate protected data scope. A successful application-state restore must not be represented as a complete media-library restore unless the referenced media scope was also recovered and verified.

Restore validation must prove that the restored backend schema is supported before the service treats the state as ready. Database backup mechanics, corruption checks, restore ordering, rollback compatibility, and recovery point/objective claims remain open until independently verified.

## Privacy and authorization boundary

Profile-owned history, favorites, ratings, queues, library membership, file observations, and extracted metadata are private application data. Storage does not create authorization by itself: service-layer authorization must still be checked before reads, writes, metadata extraction, playback, sharing, administration, or future synchronization.

GoreeCloud Sync remains a separately governed application/service capability rather than an Integral Platform System. Future Music synchronization may replicate authorized state derived from this model, but synchronization must not silently change ownership, become backup, or make device-local downloads appear present on devices that do not possess them.

## Acceptance boundary

The merged Development source establishes repository-level persistence, per-profile library authorization, source-file observation/reconciliation, profile-state, Recently Added, and bounded embedded metadata foundations with exact-head and post-merge validation. It does **not** establish:

- event-driven filesystem-change detection;
- approved sidecar metadata extraction or artwork handling;
- metadata support for every scanner-approved media container;
- canonical recording/release/source-item/playable-asset ingestion from scanned files;
- codec/container/duration/quality probing;
- library/search/profile-state/metadata HTTP APIs;
- production authentication/session handling;
- production persistence qualification or backup/restore acceptance;
- production deployment, release eligibility, overall Platform runtime conformance, or Stable status;
- Milestone 1 completion.

Every later claim requires its own exact source, tests, integration, recovery, release, and runtime evidence.
