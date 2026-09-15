# GoreeCloud Music — Repository Notes

## Current milestone

Milestone 1 — Native Multi-User Library.

## Established source foundations

- Milestone 0 established the engine-neutral Music application-state and forward-migration contract.
- Schema version 2 introduced explicit `library_memberships` for per-profile library authorization; schema version 3 adds durable `library_files` filesystem observations for incremental scan reconciliation.
- The current Development application-state backend is SQLite through Go `database/sql` and `modernc.org/sqlite`.
- Schema-v1 library owners are preserved during migration to v2 by transactionally materializing owner memberships before the stored schema version advances; the v2→v3 migration adds scanner state without changing those authorization facts.
- Profile and library state, explicit read/edit/owner membership facts, owner protections, and storage-aware health behavior are implemented and tested.
- Supported-audio library scanning is implemented for FLAC, WAV, AIFF/AIF, AAC, M4A, MP3, Opus, Ogg, and Oga extensions. Scanning does not follow symbolic links or mutate original source media.
- Library-file reconciliation records Added, Updated, Unchanged, Missing, and Restored states. Missing files are tombstoned rather than silently deleted.
- Library-file listing requires read permission; scan/reconciliation requires edit permission.
- Profile-scoped Favorites, Ratings, and Recently Played application state is implemented for authorized recordings. Recording-scoped mutations require current library read permission, and normal favorite/history queries suppress state after library authorization is revoked.
- Cross-profile tests verify default denial, explicit grants, independent rating state, and authorization-revocation suppression for the bounded profile-state paths implemented by PR #12.
- Original user media remains outside application database state.
- Reusable authentication/provider secrets remain outside ordinary application records.
- PR #7 merged the bounded persistence/authorization foundation to `main` as `564ac6a792070996dc39b103232c39b4fca95074`; post-merge CI `35017128457` and Platform Contract `35017129147` passed on that exact authoritative revision.
- PR #9 merged the bounded scanner/reconciliation foundation to `main` as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`; exact candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed CI `35019904248` and Platform Contract `35019905704`, and post-merge `main` passed CI `35020139101` and Platform Contract `35020140030`.
- PR #12 merged the bounded profile-state foundation to `main` as `168d19d07e9b32e0089a512b9e6b7964db76ece1`; exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed CI `35025974258` and Platform Contract `35025974812`, and post-merge `main` passed CI `35026228206` and Platform Contract `35026228918`.

## Open architecture decisions and implementation obligations

- Complete the remainder of Milestone 1: embedded/sidecar metadata extraction, artwork handling, canonical Recording/Release/Source Item/Playable Asset ingestion from scanned files, filesystem event watchers/change notifications, Recently Added state, complete multi-user isolation across remaining library paths, library/search/profile-state APIs, and remaining persistent first-party state.
- Complete production persistence qualification, including concurrency/load evidence, corruption handling, migration/export behavior, and Everkeep-aligned backup/restore acceptance. The merged SQLite Development backend, scanner state, and profile-state operations are not production/recovery acceptance.
- Production authentication/session integration with GoreeCloud Identity is not implemented.
- Privacy Shield, Wardveil Security, Everkeep, Manager, Mesh, and Glaze UI runtime integrations are not implemented.
- No external source provider is implemented.
- No open-source software license has been authoritatively selected; `LICENSE` is currently a rights notice.
- Production deployment/container packaging is not implemented.

These are active implementation obligations or blockers, not silently assumed defaults.
