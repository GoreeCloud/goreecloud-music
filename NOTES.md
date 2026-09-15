# GoreeCloud Music — Repository Notes

## Current milestone

Milestone 1 — Native Multi-User Library.

## Established source foundations

- Milestone 0 established the engine-neutral Music application-state and forward-migration contract.
- Schema version 2 introduced explicit `library_memberships` for per-profile library authorization; schema version 3 added durable `library_files` filesystem observations for incremental scan reconciliation; schema version 4 adds durable `library_file_metadata` bound to exact scanner observation snapshots.
- The current Development application-state backend is SQLite through Go `database/sql` and `modernc.org/sqlite`.
- Schema-v1 library owners are preserved during migration to v2 by transactionally materializing owner memberships before the stored schema version advances; the v2→v3 migration adds scanner state without changing those authorization facts; the v3→v4 migration adds per-file metadata state without mutating source media or creating canonical recording identity.
- Profile and library state, explicit read/edit/owner membership facts, owner protections, and storage-aware health behavior are implemented and tested.
- Supported-audio library scanning is implemented for FLAC, WAV, AIFF/AIF, AAC, M4A, MP3, Opus, Ogg, and Oga extensions. Scanning does not follow symbolic links or mutate original source media.
- Library-file reconciliation records Added, Updated, Unchanged, Missing, and Restored states. Missing files are tombstoned rather than silently deleted.
- Library-file listing requires read permission; scan/reconciliation requires edit permission.
- Bounded read-only embedded metadata extraction is implemented for MP3 ID3v2.3/v2.4 and FLAC Vorbis comments. The metadata layer normalizes title, artist, album, album artist, genre, date/year, track number, and disc number without inferring canonical equivalence.
- Metadata extraction requires library edit permission; metadata retrieval requires library read permission. Stored metadata becomes non-current when scanner size/mtime facts change or the source is tombstoned missing.
- Metadata source access is rooted with Go `os.Root`, rejects symbolic-link components, requires a regular file, and verifies scanner size/mtime both before parsing and again before persistence.
- Explicit-identity ingestion is implemented for a current scanner/metadata observation when the caller supplies validated canonical Recording and Release IDs. Source Item and Playable Asset IDs are deterministically derived from library plus scanner file identity; exact retries are idempotent and conflicting identity/source/path bindings fail closed.
- Bounded MP3/FLAC media format probing is implemented for already-materialized playable assets. MP3 requires actual MPEG Layer III frame evidence; FLAC requires the native `fLaC` signature and valid first STREAMINFO block shape. Only `codec` and `container` are persisted, after revalidating authorization and source freshness.
- Profile-scoped Favorites, Ratings, and Recently Played application state is implemented for authorized recordings. Recording-scoped mutations require current library read permission, and normal favorite/history queries suppress state after library authorization is revoked.
- An authorization-scoped Recently Added query is implemented over canonical recording `added_at` state. It returns only recordings from libraries the requesting profile can currently read, uses deterministic newest-first ordering, and suppresses revoked-library results without deleting underlying library state.
- Cross-profile tests verify default denial, explicit grants, independent profile state, authorization-revocation suppression, and the bounded Recently Added isolation paths implemented through PR #14.
- Original user media remains outside application database state.
- Reusable authentication/provider secrets remain outside ordinary application records.
- PR #7 merged the bounded persistence/authorization foundation to `main` as `564ac6a792070996dc39b103232c39b4fca95074`; post-merge CI `35017128457` and Platform Contract `35017129147` passed on that exact authoritative revision.
- PR #9 merged the bounded scanner/reconciliation foundation to `main` as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`; exact candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed CI `35019904248` and Platform Contract `35019905704`, and post-merge `main` passed CI `35020139101` and Platform Contract `35020140030`.
- PR #12 merged the bounded profile-state foundation to `main` as `168d19d07e9b32e0089a512b9e6b7964db76ece1`; exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed CI `35025974258` and Platform Contract `35025974812`, and post-merge `main` passed CI `35026228206` and Platform Contract `35026228918`.
- PR #14 merged the bounded Recently Added query foundation to `main` as `c20acc1a92ee71ee4173468d97a5980ccf396013`; exact candidate `8058489efbaafcfd95fbe1b6144a4eb6bc909c6d` passed CI `35028058553` and Platform Contract `35028059131`, and post-merge `main` passed CI `35028245616` and Platform Contract `35028246120`.
- PR #16 merged the bounded embedded-metadata foundation to `main` as `bc90686a18338afef650b73357f454d1be541afb`; exact candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed CI `35030658032` and Platform Contract `35030658657`, and post-merge `main` passed CI `35030857299` and Platform Contract `35030858110`.
- PR #18 merged the bounded explicit-identity materialization foundation to `main` as `f365047f33f732afb1e274cd1d4b4b401271c32a`; exact candidate `b46d2628a7fd86db79dc2a0be17945908166a8fd` passed CI `35032631291` and Platform Contract `35032632700`, and post-merge `main` passed CI `35032841324` and Platform Contract `35032841752`.
- PR #19 merged the bounded MP3/FLAC codec-container probing foundation to current `main` as `5e59466da23e5800f13f19e377a9e852631276ef`; exact candidate `d3eb0dcf31dce6b1802f0a1d979d628e15ee7a7f` passed CI `35034688612` and Platform Contract `35034689104`, and post-merge `main` passed CI `35034883077` and Platform Contract `35034883624`.

## Open architecture decisions and implementation obligations

- Complete the remainder of Milestone 1: approved sidecar metadata extraction, embedded-tag/container formats beyond the current MP3/FLAC subset, artwork handling, automatic Recording/Release matching and duplicate-equivalence policy, broader/background ingestion, duration/bitrate/sample-rate/channel/bit-depth probing, ReplayGain, media hashes/integrity, broader format probing, filesystem event watchers/change notifications, complete multi-user isolation across remaining library paths, library/search/profile-state APIs, and remaining persistent first-party state.
- Recently Added storage/query foundation is merged, but Home/user-facing Recently Added experiences and API exposure remain separate pending work.
- Complete production persistence qualification, including concurrency/load evidence, corruption handling, migration/export behavior, and Everkeep-aligned backup/restore acceptance. The merged SQLite Development backend, scanner/metadata state, explicit materialization, format probing, profile-state operations, and Recently Added query are not production/recovery acceptance.
- Production authentication/session integration with GoreeCloud Identity is not implemented.
- Privacy Shield, Wardveil Security, Everkeep, Manager, Mesh, and Glaze UI runtime integrations are not implemented.
- No external source provider is implemented.
- No open-source software license has been authoritatively selected; `LICENSE` is currently a rights notice.
- Production deployment/container packaging is not implemented.

These are active implementation obligations or blockers, not silently assumed defaults.
