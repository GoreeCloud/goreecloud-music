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

PR #7 established the current Development SQLite backend through Go `database/sql` and `modernc.org/sqlite`, including:

- persisted profiles and libraries;
- schema version 2 with explicit `library_memberships`;
- atomic owner membership at library creation;
- explicit read, edit, and owner permission evaluation;
- fail-closed owner mutation protections;
- schema-v1 owner-authorization preservation during the v1 → v2 migration;
- absolute library-root validation;
- storage-aware `/healthz` behavior;
- persistence and authorization tests across reopen/migration paths.

PR #7 merged as `564ac6a792070996dc39b103232c39b4fca95074`; post-merge Music CI `35017128457` and Platform Contract `35017129147` passed.

### Source-preserving library-file scanning and reconciliation

PR #9 advanced the Development schema to version 3 and established the filesystem-observation layer required before canonical media ingestion:

- durable `library_files` observations scoped to a library;
- supported-audio discovery for the initial approved extension set;
- no symbolic-link traversal and rejection of a symbolic-link library root;
- no source-file mutation during discovery;
- library-relative path validation that rejects root escape;
- deterministic Added, Updated, Unchanged, Missing, and Restored reconciliation;
- missing-file tombstones rather than silent identity deletion;
- read permission for listing file state and edit permission for scanning/reconciliation;
- v2 → v3 migration tests that preserve existing authorization state.

PR #9 merged as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`; post-merge Music CI `35020139101` and Platform Contract `35020140030` passed.

### Profile-scoped Favorites, Ratings, and Recently Played

PR #12 implemented bounded application-state operations over existing `favorites`, `ratings`, and `play_history` entities:

- profile-owned Favorite set/clear and bounded listing;
- profile-owned 0–100 Rating set/read/clear;
- profile-owned Recently Played event recording and bounded recent-history listing;
- current library-read authorization validation for recording-scoped profile state;
- membership-filtered Favorite/Recently Played retrieval after access revocation;
- cross-profile isolation for the bounded implemented paths;
- validation, race-test, and build coverage.

PR #12 merged as `168d19d07e9b32e0089a512b9e6b7964db76ece1`; exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed Music CI `35025974258` and Platform Contract `35025974812`, and post-merge `main` passed Music CI `35026228206` and Platform Contract `35026228918`.

### Authorization-scoped Recently Added

PR #14 implemented a bounded Recently Added library view over existing canonical recording application state:

- `RecentlyAddedForProfile` queries `recordings.added_at` without a schema change;
- newest-first deterministic ordering;
- recording/library identity, title, artist, and added timestamp output;
- current `library_memberships` filtering for the requesting profile;
- revocation/inaccessible-library suppression without mutating underlying canonical recording state;
- validation and isolation coverage for ordering, grants, revocation, malformed profile IDs, and result limits.

PR #14 merged as `c20acc1a92ee71ee4173468d97a5980ccf396013`; exact candidate `8058489efbaafcfd95fbe1b6144a4eb6bc909c6d` passed Music CI `35028058553` and Platform Contract `35028059131`, and post-merge `main` passed Music CI `35028245616` and Platform Contract `35028246120`.

### Embedded metadata extraction

PR #16 advanced the Development schema to version 4 and established a bounded, read-only embedded metadata layer over scanner observations:

- durable `library_file_metadata` keyed by scanner `file_id`;
- MP3 ID3v2.3/v2.4 textual metadata extraction;
- FLAC Vorbis Comment textual metadata extraction;
- normalized title, artist, album, album artist, genre, date/year text, track-number text, and disc-number text;
- library edit authorization for extraction/persistence and library read authorization for retrieval;
- source freshness verification against scanner size/mtime before parsing and again immediately before persistence;
- stale-state reporting when the latest scanner observation changes or the file is tombstoned missing;
- traversal-resistant rooted reads through Go `os.Root` plus explicit symbolic-link component rejection;
- bounded parser limits and malformed-input rejection;
- schema-v3 to schema-v4 migration coverage that preserves existing library/file observations.

PR #16 deliberately does not convert extracted tags into canonical Recording/Release identity and does not rewrite source media. Exact candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed Music CI `35030658032` and Platform Contract `35030658657`. It merged as signed `main` commit `bc90686a18338afef650b73357f454d1be541afb`, which passed post-merge Music CI `35030857299` and Platform Contract `35030858110`.

## Storage and media boundary

Original music files remain in GoreeCloud-controlled library storage. SQLite stores application state, references/observations, and bounded extracted metadata facts only; it does not embed original media bytes. Media-file loss and database-state loss remain separate recovery domains. Reusable authentication credentials and provider secrets remain outside ordinary Music application-state records.

The deterministic library-file identifier used by the scanner is a reconciliation identity derived from library identity and relative path. It is not a content hash and does not establish canonical Recording identity. Extracted metadata is tied to that scanner observation and is likewise not canonical Recording or Release identity.

Favorites, Ratings, Recently Played, Recently Added, and extracted per-file metadata use application-state records/entities; their implementation does not itself qualify canonical media ingestion or production persistence.

## Authorization boundary

Durable library membership supplies application authorization facts but does not replace higher-level GoreeCloud authority. API/service operations must continue to validate the requesting profile, operation, purpose, applicable Privacy Shield authorization, Wardveil requirements, and future GoreeCloud Identity/session authority.

Queue persistence, profile state, Recently Added results, extracted metadata, or previously observed file state do not grant perpetual access. Authorization must be rechecked at the operation/playback/retrieval boundary where required.

## Current status boundary

The verified foundations do **not** establish:

- approved sidecar metadata extraction;
- artwork discovery/processing;
- embedded metadata support for every scanner-approved container;
- canonical Recording, Release, Source Item, or Playable Asset ingestion from scanned files;
- codec/container/duration/quality probing;
- background metadata refresh orchestration;
- event-driven filesystem watchers/change notifications;
- Home/user-facing Recently Added or metadata integration;
- library/search/profile-state/metadata product APIs;
- actual audio streaming/transcoding/playback sessions;
- production authentication/session integration;
- real external provider adapters;
- production persistence, corruption, backup/restore, or Everkeep acceptance;
- user-facing web, Android, Linux, iOS, automotive, or television clients;
- current-Stable Glaze UI acceptance;
- release eligibility or Stable qualification.

Those remain later Milestone 1 or subsequent milestone obligations and must not be inferred from the persistence/scanner/profile-state/Recently Added/metadata foundations.
