# Milestone 1 — Native Multi-User Library

## Status

Development implementation is active on `main`. PR #7 merged the bounded SQLite persistence and per-user library-authorization foundation as `564ac6a792070996dc39b103232c39b4fca95074`. PR #9 then merged the bounded source-preserving library-scanning and filesystem-reconciliation foundation as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`. PR #12 subsequently merged the bounded authorization-scoped Favorites, Ratings, and Recently Played persistence foundation as `168d19d07e9b32e0089a512b9e6b7964db76ece1`. PR #14 then merged the bounded authorization-scoped Recently Added query foundation as `c20acc1a92ee71ee4173468d97a5980ccf396013`. PR #16 now adds the bounded embedded MP3/FLAC metadata extraction and durable metadata-state foundation as signed `main` commit `bc90686a18338afef650b73357f454d1be541afb`.

PR #9's exact candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed Music CI `35019904248` and Platform Contract `35019905704`. Its merged revision passed post-merge Music CI `35020139101` and Platform Contract `35020140030`.

PR #12's exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed Music CI `35025974258` and Platform Contract `35025974812`. The merged `main` revision passed post-merge Music CI `35026228206` and Platform Contract `35026228918`.

PR #14's exact candidate `8058489efbaafcfd95fbe1b6144a4eb6bc909c6d` passed Music CI `35028058553` and Platform Contract `35028059131`. The merged `main` revision passed post-merge Music CI `35028245616` and Platform Contract `35028246120`.

PR #16's exact candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed Music CI `35030658032` and Platform Contract `35030658657`. The signed merged `main` revision `bc90686a18338afef650b73357f454d1be541afb` passed post-merge Music CI `35030857299` and Platform Contract `35030858110`.

This record describes implemented Milestone 1 foundations and remaining work. It is not Milestone 1 completion, release evidence, production persistence acceptance, recovery acceptance, or Stable qualification.

## Objective

Build the first durable native GoreeCloud Music library service on top of the accepted Milestone 0 domain, routing, authorization-request, and engine-neutral storage contracts.

Milestone 1 must establish a multi-user library without turning the database into media storage, weakening user isolation, or allowing a storage engine to become the Music domain authority.

## Current bounded implementation

Merged PR #7 provides:

- SQLite application-state persistence through Go `database/sql` and `modernc.org/sqlite`;
- forward-only schema migration from version 1 to version 2;
- explicit `library_memberships` for per-profile authorization;
- transactionally preserved schema-v1 library-owner authorization during migration to v2;
- profile and library persistence;
- atomic owner membership at library creation;
- read, edit, and owner permission evaluation;
- fail-closed owner mutation rules;
- absolute filesystem root validation for libraries;
- persisted library visibility across service/database reopen;
- storage-aware `/healthz` behavior;
- a configurable local Development database path;
- tests covering fresh migration, schema-v1 → v2 owner preservation, persistence, default denial, explicit grant, owner protections, root-path validation, and SQLite safety settings.

Merged PR #9 adds:

- forward-only schema migration from version 2 to version 3;
- durable `library_files` observations containing library-relative path, size, modification time, first/last seen state, and an optional missing tombstone;
- supported-audio discovery for FLAC, WAV, AIFF/AIF, AAC, M4A, MP3, Opus, Ogg, and Oga extensions;
- symbolic-link avoidance and rejection of a symbolic-link library root;
- scan behavior that does not write, rename, move, delete, or rewrite source media;
- validated library-relative paths that may not escape the configured root;
- deterministic Added, Updated, Unchanged, Missing, and Restored reconciliation states;
- `missing_since` tombstones instead of silent database identity deletion when a source file disappears;
- read permission for file-state queries and edit permission for scan/reconciliation mutations;
- schema-v2 → v3 migration tests demonstrating existing authorization remains intact;
- source-preservation, discovery, isolation, unsafe-path, modification, missing-file, and restoration tests.

Merged PR #12 adds:

- profile-owned Favorite set/clear operations and bounded favorite listing;
- profile-owned 0–100 Rating set/read/clear operations;
- profile-owned Recently Played event persistence and bounded recent-history listing;
- current library-read authorization checks before recording-scoped profile-state mutations and direct rating reads;
- membership-filtered Favorites and Recently Played queries so revoked library access does not remain visible through normal Music surfaces;
- independent profile state when multiple profiles can read the same library;
- validation for result limits, rating range, playback timestamps, and playback durations;
- tests covering default denial, explicit library access grants, cross-profile state isolation, authorization-revocation suppression, validation behavior, race testing, and service build.

PR #12 deliberately does not add a new schema version because `favorites`, `ratings`, and `play_history` already exist in the current application-state schema. Its tests seed canonical recording rows directly because recording ingestion remains a separate Milestone 1 obligation.

Merged PR #14 adds:

- a bounded `RecentlyAddedForProfile` query over canonical `recordings.added_at` application state;
- newest-first ordering with deterministic recording-ID tie breaking;
- returned recording ID, library ID, title, artist, and added timestamp;
- current `library_memberships` filtering for the requesting profile;
- suppression of inaccessible or revoked-library recordings without rewriting/deleting underlying library state;
- tests covering ordering, cross-library isolation, explicit access grants, authorization revocation, result bounds, malformed profile IDs, race testing, and service build.

PR #14 deliberately does not add a new schema version or dependency. Its tests seed canonical recording rows directly because canonical recording ingestion from verified scanned media remains a separate Milestone 1 obligation. The query foundation does not itself implement Home UI or HTTP API exposure.

Merged PR #16 adds:

- forward-only schema migration from version 3 to version 4;
- durable `library_file_metadata` keyed by scanner `file_id`, storing normalized embedded metadata plus the scanner size/mtime snapshot used for extraction;
- bounded read-only MP3 ID3v2.3/v2.4 extraction and FLAC Vorbis Comment extraction;
- normalized title, artist, album, album artist, genre, date/year text, track-number text, and disc-number text without treating those values as canonical Recording/Release identity;
- library edit permission for extraction/replacement and library read permission for metadata retrieval;
- `Current` metadata state derived from comparison with the latest `library_files` observation and its missing tombstone state;
- pre-read and post-read size/mtime verification on the same opened source descriptor so metadata is not persisted against a source that changed during parsing;
- traversal-resistant source access rooted with Go `os.Root`, while retaining the stricter Music policy that rejects symbolic-link path components;
- bounded tag/block/comment limits and malformed encoding/container rejection;
- schema-v3 → v4 migration regression coverage preserving existing library/file state;
- tests for authorization, stale scanner observations, missing files, symlink substitution, unsupported containers, parser correctness, race testing, and service build.

PR #16 deliberately does not create canonical Recording, Release, Source Item, or Playable Asset entities from extracted tags. It also does not implement artwork, approved sidecars, additional container formats, codec probing, background refresh orchestration, public APIs, or user-facing metadata UI.

## Data boundary

Original music files remain in GoreeCloud-controlled library storage. SQLite stores application state, file references/observations, and bounded extracted metadata facts only. Media-file loss and database loss are therefore separate recovery domains.

A `library_files` row is a filesystem observation, not a canonical Recording identity, content fingerprint, or playable-asset qualification. A `library_file_metadata` row is metadata tied to that observation snapshot; it is likewise not canonical Recording/Release identity. File extension recognition does not prove actual codec validity.

Favorites, ratings, ordinary listening history, Recently Added views, and extracted per-file metadata are application state derived from authorized library records. They do not replace operational telemetry, security/audit records, or Privacy Shield purpose/retention controls where those controls apply.

Reusable secrets and external-provider credentials are outside ordinary Music application-state records.

## Authorization boundary

Library membership is explicit and fail-closed. A profile that has no membership receives no library permission. Creating a library creates only the owner's membership. Granting another profile access requires an explicit operation.

The schema-v1 → v2 migration preserves existing library ownership by materializing owner memberships in the same transaction that creates `library_memberships`, before the stored schema version advances. The schema-v2 → v3 migration adds file observations without altering membership authority. The schema-v3 → v4 migration adds metadata state without changing existing library or file ownership facts.

Reading scanner state requires library read permission. Running a scan or directly reconciling observations requires library edit permission. Reading extracted metadata requires library read permission; extracting or replacing durable metadata requires library edit permission.

Favorites, ratings, and Recently Played mutations require the profile to retain current read permission to the recording's library. Favorite/history listing joins current library membership so revoked library access suppresses stale state from normal user-facing retrieval. A direct rating read also revalidates current recording/library authorization.

Recently Added retrieval joins current library membership for the requesting profile. Records from inaccessible or revoked libraries are excluded without deleting the underlying canonical recording state.

The backend provides durable authorization facts; it does not replace higher-level authorization. Service/API operations must continue to validate the requesting profile, action, purpose, Privacy Shield authorization where applicable, Wardveil requirements, and future Identity authority.

## Remaining Milestone 1 work

Milestone 1 remains open. Required work includes:

- approved sidecar metadata extraction and additional embedded metadata formats where required;
- album/artist artwork discovery, storage policy, and reconciliation;
- recording, release, source-item, and playable-asset ingestion from verified scanned media;
- codec/container probing beyond filename-extension discovery;
- bulk/background metadata refresh orchestration;
- filesystem watchers or another approved event-driven change-detection mechanism where appropriate;
- Home/user-facing Recently Added integration and API exposure;
- higher-level Favorites, Ratings, Recently Played, and metadata service/API integration;
- complete multi-user isolation tests across every remaining library query/mutation path;
- library and local search APIs;
- migration/export and recovery controls;
- Everkeep-aligned backup/restore design and tested restoration;
- production authentication/session integration;
- observability, performance, corruption/failure, and concurrency acceptance.

## Verification boundary

PR #7's merged revision `564ac6a792070996dc39b103232c39b4fca95074` passed post-merge Music CI `35017128457` and Platform Contract `35017129147`.

PR #9's exact source candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed Music CI `35019904248`, including exact-source verification, formatting, vet, race tests, and service build, and Platform Contract `35019905704`. It merged as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`, which passed post-merge Music CI `35020139101` and Platform Contract `35020140030`.

PR #12's exact source candidate `2721d2b7660763534396be4017ab9c18e47078da` passed Music CI `35025974258`, including exact-source verification, formatting, vet, race tests, and service build, and Platform Contract `35025974812`. It merged as `168d19d07e9b32e0089a512b9e6b7964db76ece1`, which passed post-merge Music CI `35026228206` and Platform Contract `35026228918`.

PR #14's exact source candidate `8058489efbaafcfd95fbe1b6144a4eb6bc909c6d` passed Music CI `35028058553`, including exact-source verification, formatting, vet, race tests, and service build, and Platform Contract `35028059131`. It merged as `c20acc1a92ee71ee4173468d97a5980ccf396013`, which passed post-merge Music CI `35028245616` and Platform Contract `35028246120`.

PR #16's exact source candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed Music CI `35030658032`, including exact-source verification, formatting, vet, race tests, and service build, and Platform Contract `35030658657`. It merged as signed `main` commit `bc90686a18338afef650b73357f454d1be541afb`, which passed post-merge Music CI `35030857299` and Platform Contract `35030858110`.

Those results establish only the checks executed against the merged Development source. They do not establish production deployment, runtime Platform-System acceptance, backup/restore acceptance, current-Stable Glaze UI acceptance, release eligibility, or Stable status.

Any later source change requires validation appropriate to that new exact revision.
