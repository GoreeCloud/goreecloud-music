# GoreeCloud Music — Repository Specifications

**Repository version:** `0.1.0-dev.1`  
**Lifecycle:** Active Development  
**Current milestone:** Milestone 1 — Native Multi-User Library

This file records repository-coupled implementation specifications. The authoritative product scope and planned capability direction are maintained in `GoreeCloud/Projects/Project Specification — Music.md`.

## Native architecture

GoreeCloud Music is original GoreeCloud-owned software. Whole-product Navidrome or other third-party application architecture is not the implementation foundation. OpenSubsonic/Navidrome compatibility may be added later through bounded interoperability layers.

## Established Milestone 0 contracts

The merged Milestone 0 foundation established:

- stable domain identity types for Recording, Release, Source Item, Playable Asset, Queue Item, Profile, and Library;
- source identity independent from availability state;
- Exact, Equivalent, Alternate, and Unknown match confidence;
- provider-neutral adapter interfaces;
- application authorization request/decision contracts;
- deterministic playback route selection with exact-match protection;
- stable queue-item identity and source/route context;
- an engine-neutral application-state schema and forward-only migration model;
- minimal native service identity/health endpoints;
- an OpenAPI Development contract that separates implemented paths from planned domains;
- truthful machine-readable Platform-System state;
- baseline tests and exact-source GitHub Actions validation.

## Verified Milestone 1 foundations

### SQLite persistence and library authorization

PR #7 established the current Development SQLite backend through Go `database/sql` and `modernc.org/sqlite`, including persisted profiles and libraries, schema version 2 with explicit `library_memberships`, atomic owner membership, explicit read/edit/owner permission evaluation, fail-closed owner mutation protections, schema-v1 owner-authorization preservation, absolute library-root validation, storage-aware `/healthz`, and persistence/authorization tests. PR #7 merged as `564ac6a792070996dc39b103232c39b4fca95074`; post-merge Music CI `35017128457` and Platform Contract `35017129147` passed.

### Source-preserving library-file scanning and reconciliation

PR #9 advanced the Development schema to version 3 and established durable `library_files`, supported-audio discovery, no symbolic-link traversal, source-preserving scans, root-escape rejection, deterministic Added/Updated/Unchanged/Missing/Restored reconciliation, missing-file tombstones, read/edit authorization boundaries, and v2→v3 migration coverage. PR #9 merged as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`; post-merge Music CI `35020139101` and Platform Contract `35020140030` passed.

### Profile-scoped Favorites, Ratings, and Recently Played

PR #12 implemented bounded operations over existing `favorites`, `ratings`, and `play_history` state with current library authorization, cross-profile isolation, revocation suppression, validation, race tests, and build coverage. PR #12 merged as `168d19d07e9b32e0089a512b9e6b7964db76ece1`; exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed Music CI `35025974258` and Platform Contract `35025974812`, and post-merge `main` passed Music CI `35026228206` and Platform Contract `35026228918`.

### Authorization-scoped Recently Added

PR #14 implemented `RecentlyAddedForProfile` over existing canonical `recordings.added_at` state with newest-first deterministic ordering, current membership filtering, revocation suppression, and focused validation/isolation tests. PR #14 merged as `c20acc1a92ee71ee4173468d97a5980ccf396013`; exact candidate `8058489efbaafcfd95fbe1b6144a4eb6bc909c6d` passed Music CI `35028058553` and Platform Contract `35028059131`, and post-merge `main` passed Music CI `35028245616` and Platform Contract `35028246120`.

### Embedded metadata extraction and persistence

PR #16 advanced the Development schema to version 4 and established a bounded metadata layer between scanner observations and future canonical media ingestion:

- `library_file_metadata` keyed to scanner `file_id`;
- read-only MP3 ID3v2.3/v2.4 text extraction;
- read-only FLAC Vorbis Comment extraction;
- normalized title, artist, album, album artist, genre, date/year, track-number, and disc-number facts;
- exact source size/mtime snapshot binding and current/stale evaluation;
- edit authorization for extraction/persistence and read authorization for retrieval;
- rooted source access using Go `os.Root`, explicit symlink rejection, regular-file validation, and pre/post-parse scanner snapshot verification;
- bounded malformed-input handling with no canonical identity creation on parser failure;
- schema-v3 → v4, metadata persistence, authorization, staleness, unsupported-container, and symlink-substitution tests.

PR #16 exact candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed Music CI `35030658032` and Platform Contract `35030658657`. It merged as `bc90686a18338afef650b73357f454d1be541afb`; post-merge Music CI `35030857299` and Platform Contract `35030858110` passed.

## Storage and media boundary

Original music files remain in GoreeCloud-controlled library storage. SQLite stores application state, references/observations, and normalized metadata facts only; it does not embed original media bytes. Media-file loss and database-state loss remain separate recovery domains. Reusable authentication credentials and provider secrets remain outside ordinary Music application-state records.

The deterministic library-file identifier used by the scanner is a reconciliation identity derived from library identity and relative path. It is not a content hash and does not establish canonical Recording identity. `library_file_metadata` is similarly a bounded extraction record tied to the scanner observation; tag facts do not silently become canonical Recording/Release identity.

Favorites, Ratings, Recently Played, and Recently Added use existing application-state records/entities; their implementation does not itself qualify canonical media ingestion or production persistence.

## Authorization boundary

Durable library membership supplies application authorization facts but does not replace higher-level GoreeCloud authority. API/service operations must continue to validate the requesting profile, operation, purpose, applicable Privacy Shield authorization, Wardveil requirements, and future GoreeCloud Identity/session authority.

Queue persistence, profile state, Recently Added results, previously observed file state, or extracted metadata do not grant perpetual access. Authorization must be rechecked at the operation/playback/retrieval boundary where required.

## Current status boundary

The verified foundations do **not** establish:

- approved sidecar metadata extraction;
- embedded metadata support beyond the bounded MP3 ID3v2.3/v2.4 and FLAC Vorbis Comment subset;
- artwork discovery/processing;
- canonical Recording, Release, Source Item, or Playable Asset ingestion from scanned files;
- codec/container probing beyond filename-extension discovery;
- event-driven filesystem watchers/change notifications;
- Home/user-facing Recently Added integration;
- library/search/profile-state product APIs;
- actual audio streaming/transcoding/playback sessions;
- production authentication/session integration;
- real external provider adapters;
- production persistence, corruption, backup/restore, or Everkeep acceptance;
- user-facing web, Android, Linux, iOS, automotive, or television clients;
- current-Stable Glaze UI acceptance;
- release eligibility or Stable qualification.

Those remain later Milestone 1 or subsequent milestone obligations and must not be inferred from the persistence/scanner/metadata/profile-state/Recently Added foundations.
