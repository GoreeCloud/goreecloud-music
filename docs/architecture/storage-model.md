# GoreeCloud Music Storage Model and Migration Contract

## Status

The accepted Milestone 0 foundation remains an engine-neutral durable application-state contract. Draft Milestone 1 PR #7 now adds a concrete SQLite Development backend while preserving that logical contract and its replaceability boundary.

This document distinguishes the accepted logical storage authority from the current Draft backend candidate. A backend implementation does not become production-authoritative merely because it satisfies the logical schema or passes repository CI.

## Purpose

GoreeCloud Music separates durable application state from original user media. Milestone 0 established the logical state model and forward-only migration contract before a database engine was selected. Milestone 1 may add a concrete backend only when it preserves the same ownership, privacy, portability, recovery, and migration boundaries.

The engine-neutral executable contract lives in `internal/storage`. The current Draft SQLite implementation lives in `internal/storage/sqlite`.

## Storage boundary

The application-state backend may persist identities, metadata, ownership and authorization references, queue state, ratings, favorites, history, migration metadata, and durable media references. It must not become the authoritative storage location for original user-owned audio files.

Original media remains in GoreeCloud-controlled library storage and is referenced by durable identity and path. Losing or replacing the application-state backend must not by itself imply loss of the original media library.

Reusable provider credentials, private keys, authentication secrets, and equivalent sensitive credentials remain outside this storage model. A production implementation must use the applicable secret or identity authority rather than embedding reusable secrets in ordinary application records.

## Current logical schema — version 2 candidate

Draft PR #7 advances `internal/storage.CurrentSchema()` from schema version `1` to `2` by adding explicit per-profile library membership state. The candidate contains these engine-neutral entities:

| Entity | Ownership boundary | Purpose |
| --- | --- | --- |
| `schema_metadata` | system | schema version and migration metadata |
| `profiles` | profile | Music profile application state without authentication credentials |
| `libraries` | library | library identity, ownership, and source-root metadata |
| `library_memberships` | library | per-profile library authorization and role state |
| `recordings` | library | canonical recording identity and library metadata |
| `releases` | library | release and edition identity |
| `source_items` | library | provider-specific source identity mappings |
| `playable_assets` | library | encoded asset metadata, quality, path reference, and availability facts; never original media bytes |
| `queues` | profile | profile-owned durable queue identity and revision state |
| `queue_items` | profile | queue ordering, recording/source intent, selected route, and route reason |
| `favorites` | profile | profile-scoped favorite state |
| `ratings` | profile | profile-scoped rating state |
| `play_history` | profile | profile-scoped listening history subject to privacy and retention controls |

The logical model remains independent from SQLite. Future backend migration must preserve these product identities and authorization boundaries rather than exposing backend-local identity as the Music domain model.

## Migration contract

The migration catalog is forward-only and versioned independently from the GoreeCloud Music software release number.

The accepted initial migration is `0001-core-application-state`, moving an uninitialized backend from schema version `0` to `1`.

Draft PR #7 adds `0002-library-memberships`, moving schema version `1` to `2` and creating the explicit per-profile library authorization relation. The SQLite implementation performs the table creation and owner-membership backfill in the same transaction, preserving existing schema-v1 library owners before the stored schema version advances. A regression test constructs real schema-v1 SQLite state, upgrades it to v2, and verifies that owner authorization and library visibility survive the migration.

The catalog validator fails closed when migration identifiers are missing or duplicated, versions are non-contiguous or non-forward, changes are empty or unsupported, a change references an unknown entity, an entity is created more than once, the current schema is not fully represented, or the catalog does not end at the declared current schema version.

Downgrades are not implicitly supported. Rollback must be designed around data safety, backup and restoration, version compatibility, and explicit recovery evidence rather than inferred by reversing migrations.

## Backend contract

A concrete backend implements the minimal `storage.Backend` boundary:

- report its current schema version;
- apply one declared migration atomically;
- update stored schema version only after the migration succeeds.

The shared migration runner verifies the final backend version and fails when the requested target is not reached. A backend must not silently coerce unsupported schema, skip unknown migrations, or mark a failed migration as applied.

## Milestone 1 SQLite Development candidate

Draft PR #7 selects SQLite as the concrete GoreeCloud Music application-state backend for the current Development candidate, using Go `database/sql` with `modernc.org/sqlite` v1.58.0. The dependency graph is explicitly pinned, including the SQLite module's required matching `modernc.org/libc` v1.75.6 dependency.

The selection is bounded to application-state persistence. It does not make SQLite a permanent GoreeCloud identity or prevent later migration when verified requirements justify another backend.

The candidate configures:

- foreign-key enforcement;
- WAL journaling;
- full synchronous durability;
- SQLite defensive mode;
- a bounded busy timeout;
- transactional schema migrations;
- schema-v1 to schema-v2 owner-authorization preservation;
- STRICT tables for the current physical schema;
- a local Development database path configurable through `GOREECLOUD_MUSIC_DB`.

The candidate also implements profile and library persistence, atomic owner-membership creation, explicit read/edit/owner membership checks, absolute library-root validation, persistent membership state across reopen, and storage-aware health reporting.

Owner authority fails closed: an owner cannot be silently downgraded and owner permission cannot be granted to another profile without a separately designed ownership-transfer operation.

## Why SQLite is appropriate for this bounded candidate

The current self-hosted Music service is a single application authority rather than a distributed multi-writer database cluster. SQLite provides transactional relational state, explicit schema/migrations, indexes and constraints, a small operational footprint, file-level portability, and simple self-hosted deployment while keeping original media independent from the database.

This choice remains subject to workload, concurrency, recovery, corruption, observability, and production-acceptance evidence. If later Music requirements exceed the verified operating envelope, the engine-neutral storage contract permits a controlled migration to another backend without redefining Music identities or authorization semantics.

## Everkeep and recovery boundary

Selecting and validating a Development backend is not backup or recovery acceptance. Before release, the SQLite application-state database and required schema/migration metadata must become part of an Everkeep-aligned backup and restore plan with tested restoration evidence.

Original media remains a separate protected data scope. A successful application-state restore must not be represented as a complete media-library restore unless the referenced media scope was also recovered and verified.

Restore validation must prove that the restored backend schema is supported before the service treats the state as ready. Database backup mechanics, corruption checks, restore ordering, rollback compatibility, and recovery point/objective claims remain open until independently verified.

## Privacy and authorization boundary

Profile-owned history, favorites, ratings, queues, and library membership are private application data. Storage does not create authorization by itself: service-layer authorization must still be checked before reads, writes, playback, sharing, administration, or future synchronization.

GoreeCloud Sync remains a separately governed application/service capability rather than an Integral Platform System. Future Music synchronization may replicate authorized state derived from this model, but synchronization must not silently change ownership, become backup, or make device-local downloads appear present on devices that do not possess them.

## Acceptance boundary

The current Draft candidate establishes only a verified repository-level persistence and per-profile library-authorization foundation when its exact-head CI and Platform Contract checks are green. It does **not** establish:

- authoritative merged Milestone 1 source until PR #7 is reviewed and merged;
- incremental library scanning or filesystem-change detection;
- metadata extraction or artwork handling;
- recording/release ingestion into the native library;
- favorites, ratings, recently-added, or recently-played service operations;
- library/search HTTP APIs;
- production authentication/session handling;
- production backup/restore acceptance;
- production deployment, release eligibility, overall Platform conformance, or Stable status;
- Milestone 1 completion.

Every later claim requires its own exact source, tests, integration, recovery, release, and runtime evidence.
