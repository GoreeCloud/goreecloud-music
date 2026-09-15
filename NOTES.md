# GoreeCloud Music — Repository Notes

## Current milestone

Milestone 1 — Native Multi-User Library.

## Established source foundations

- Milestone 0 established the engine-neutral Music application-state and forward-migration contract.
- Schema version 2 is now implemented in `internal/storage` with explicit `library_memberships` for per-profile library authorization.
- The current Development application-state backend is SQLite through Go `database/sql` and `modernc.org/sqlite`.
- Schema-v1 library owners are preserved during migration to v2 by transactionally materializing owner memberships before the stored schema version advances.
- Profile and library state, explicit read/edit/owner membership facts, owner protections, and storage-aware health behavior are implemented and tested.
- Original user media remains outside application database state.
- Reusable authentication/provider secrets remain outside ordinary application records.
- PR #7 merged this bounded Milestone 1 foundation to `main` as `564ac6a792070996dc39b103232c39b4fca95074`; post-merge CI `35017128457` and Platform Contract `35017129147` passed on that exact authoritative revision.

## Open architecture decisions and implementation obligations

- Complete the remainder of Milestone 1: incremental scanning, filesystem-change detection, metadata/artwork handling, recording/release/source-item/playable-asset ingestion, favorites/ratings/recent activity operations, multi-user isolation across library paths, and library/search APIs.
- Complete production persistence qualification, including concurrency/load evidence, corruption handling, migration/export behavior, and Everkeep-aligned backup/restore acceptance. The merged SQLite Development backend is not production/recovery acceptance.
- Production authentication/session integration with GoreeCloud Identity is not implemented.
- Privacy Shield, Wardveil Security, Everkeep, Manager, Mesh, and Glaze UI runtime integrations are not implemented.
- No external source provider is implemented.
- No open-source software license has been authoritatively selected; `LICENSE` is currently a rights notice.
- Production deployment/container packaging is not implemented.

These are active implementation obligations or blockers, not silently assumed defaults.
