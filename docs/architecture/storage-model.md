# Milestone 0 Storage Model and Migration Contract

## Status

Engine-neutral Milestone 0 architecture contract for GoreeCloud Music. It defines durable application-state boundaries and schema migration behavior without selecting or claiming a production persistence engine.

## Purpose

The Music product specification requires Milestone 0 to define a storage model and migrations before the native multi-user library begins. This contract closes that architecture gap while keeping the database/backend selection subordinate to validated library, deployment, recovery, and operational requirements.

The executable contract lives in `internal/storage`.

## Storage boundary

GoreeCloud Music separates durable **application state** from original user media.

The application-state backend may persist identities, metadata, ownership/authorization references, queue state, ratings, favorites, history, and migration metadata. It must not become the authoritative storage location for original user-owned audio files.

Original media remains in GoreeCloud-controlled library storage and is referenced by durable identity. Losing or replacing the application-state backend must not by itself imply loss of the original media library.

Reusable provider credentials, private keys, authentication secrets, and equivalent sensitive credentials are also outside this storage model. A production implementation must use the applicable secret/identity authority rather than embedding reusable secrets in ordinary application records.

## Schema version 1

`internal/storage.CurrentSchema()` currently declares schema version `1` with these engine-neutral entities:

| Entity | Ownership boundary | Purpose |
| --- | --- | --- |
| `schema_metadata` | system | schema version and migration metadata |
| `profiles` | profile | Music profile application state without authentication credentials |
| `libraries` | library | library identity, ownership, and authorization metadata |
| `recordings` | library | canonical recording identity and metadata |
| `releases` | library | release and edition identity |
| `source_items` | library | provider-specific source identity mappings |
| `playable_assets` | library | encoded asset metadata, quality, and availability facts; not original media bytes |
| `queues` | profile | profile-owned queue identity and revision state |
| `queue_items` | profile | queue ordering, recording/source intent, selected route, and route reason |
| `favorites` | profile | profile-scoped favorite state |
| `ratings` | profile | profile-scoped rating state |
| `play_history` | profile | profile-scoped listening history subject to privacy and retention controls |

The model is intentionally logical rather than SQL-, document-, or key-value-specific. A future backend may add backend-local indexes, constraints, transaction metadata, or physical layout only when they preserve this product authority and isolation model.

## Migration contract

The migration catalog is forward-only and versioned independently from the GoreeCloud Music software release number.

The initial migration is `0001-core-application-state`, moving an uninitialized backend from schema version `0` to schema version `1`.

The catalog validator fails closed when:

- migration identifiers are missing or duplicated;
- migration versions are not contiguous or do not move forward;
- a migration contains no declared changes;
- a change uses an unsupported operation kind;
- a change references an entity outside the current schema;
- the same entity is created more than once;
- the migration catalog does not account for every current schema entity;
- the catalog does not end at the declared current schema version.

Downgrades are not implicitly supported. A future downgrade or rollback mechanism must be designed explicitly around data safety, backup/restoration, and version compatibility rather than inferred by reversing forward migrations.

## Backend contract

A concrete backend implements the minimal `storage.Backend` boundary:

- report its current schema version;
- apply one declared migration.

Each backend is responsible for making an individual migration atomic for that backend and for updating its stored schema version only after the migration succeeds. The migration runner verifies the final backend version and fails when the backend did not reach the requested target.

A backend implementation must not silently coerce an unsupported schema, skip unknown migrations, or mark a failed migration as applied.

## Production engine selection remains open

Milestone 0 deliberately does **not** select PostgreSQL, SQLite, another relational database, a document database, or another persistence engine. That selection belongs to a bounded implementation decision informed by the native multi-user library and deployment model.

Selection criteria must include at least:

- multi-user isolation and concurrency;
- transactional correctness and migration safety;
- indexing/search requirements;
- backup, restore, and point-in-time recovery behavior where applicable;
- corruption detection and integrity controls;
- portability and exportability;
- container/self-hosted operational footprint;
- upgrade/rollback characteristics;
- observability and maintenance burden;
- security and secret-handling boundaries.

Choosing an engine later does not authorize changing canonical Music domain identities or weakening the storage/privacy boundaries in this contract.

## Everkeep and recovery boundary

A production storage backend must become part of an Everkeep-aligned backup and restore plan before release acceptance. The recoverable application-state scope must include schema/migration metadata required to interpret the restored state.

Original media remains a separate protected data scope. A successful application-state restore must not be represented as a complete media-library restore unless the referenced media scope was also recovered and verified.

Restore validation must prove that the restored backend schema is supported before the service treats the state as ready.

## Privacy and authorization boundary

Profile-owned history, favorites, ratings, and queue state are private application data. Storage does not create authorization: service-layer authorization must still be checked before reads, writes, playback, sharing, administration, or future synchronization.

GoreeCloud Sync remains a separately governed application/service capability rather than an Integral Platform System. Future Music synchronization may replicate authorized state derived from this model, but synchronization must not silently change ownership, become backup, or make device-local downloads appear present on devices that do not possess them.

## Acceptance boundary

This contract establishes the Milestone 0 logical storage and migration foundation. It does **not** establish:

- a production persistence engine or physical schema;
- a deployed database;
- native library scanning/indexing;
- production authentication/session handling;
- media-file storage or streaming;
- production backup/restore acceptance;
- Milestone 1 completion;
- release, deployment, production acceptance, or Stable status.

Those claims require their own exact source, tests, integration, recovery, release, and runtime evidence.
