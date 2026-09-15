# GoreeCloud Music — Repository Notes

## Current milestone

Milestone 0 — Architecture Foundation.

## Established Milestone 0 boundaries

- Engine-neutral application-state schema version 1 is defined in `internal/storage`.
- Forward-only migration catalog/planning and a backend migration contract are defined and tested.
- Original user media remains outside application database state.
- Reusable authentication/provider secrets remain outside ordinary application records.
- A production persistence/database engine remains intentionally unselected; the logical model does not claim a deployed backend.

## Open architecture decisions and implementation obligations

- Select and implement a production persistence/database backend during the bounded native-library implementation, with multi-user, transaction, migration, backup/restore, integrity, portability, and operational requirements verified.
- Production authentication/session integration with GoreeCloud Identity is not implemented.
- Privacy Shield, Wardveil Security, Everkeep, Manager, Mesh, and Glaze UI runtime integrations are not implemented.
- No external source provider is implemented.
- No open-source software license has been authoritatively selected; `LICENSE` is currently a rights notice.
- Production deployment/container packaging is not implemented.

These are active implementation obligations or blockers, not silently assumed defaults.
