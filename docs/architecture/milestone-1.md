# Milestone 1 — Native Multi-User Library

## Status

Development implementation is active on `main`. PR #7 merged the bounded SQLite persistence and per-user library-authorization foundation as `564ac6a792070996dc39b103232c39b4fca95074`. PR #9 then merged the bounded source-preserving library-scanning and filesystem-reconciliation foundation as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`. PR #12 subsequently merged the bounded authorization-scoped Favorites, Ratings, and Recently Played persistence foundation as `168d19d07e9b32e0089a512b9e6b7964db76ece1`.

PR #9's exact candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed Music CI `35019904248` and Platform Contract `35019905704`. Its merged revision passed post-merge Music CI `35020139101` and Platform Contract `35020140030`.

PR #12's exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed Music CI `35025974258` and Platform Contract `35025974812`. The merged `main` revision passed post-merge Music CI `35026228206` and Platform Contract `35026228918`.

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

## Data boundary

Original music files remain in GoreeCloud-controlled library storage. SQLite stores application state and file references/observations only. Media-file loss and database loss are therefore separate recovery domains.

A `library_files` row is a filesystem observation, not a canonical Recording identity, content fingerprint, or playable-asset qualification. File extension recognition does not prove actual codec validity.

Favorites, ratings, and ordinary listening history are profile-owned application state. They do not replace operational telemetry, security/audit records, or Privacy Shield purpose/retention controls where those controls apply.

Reusable secrets and external-provider credentials are outside ordinary Music application-state records.

## Authorization boundary

Library membership is explicit and fail-closed. A profile that has no membership receives no library permission. Creating a library creates only the owner's membership. Granting another profile access requires an explicit operation.

The schema-v1 → v2 migration preserves existing library ownership by materializing owner memberships in the same transaction that creates `library_memberships`, before the stored schema version advances. The schema-v2 → v3 migration adds file observations without altering membership authority.

Reading scanner state requires library read permission. Running a scan or directly reconciling observations requires library edit permission.

Favorites, ratings, and Recently Played mutations require the profile to retain current read permission to the recording's library. Favorite/history listing joins current library membership so revoked library access suppresses stale state from normal user-facing retrieval. A direct rating read also revalidates current recording/library authorization.

The backend provides durable authorization facts; it does not replace higher-level authorization. Service/API operations must continue to validate the requesting profile, action, purpose, Privacy Shield authorization where applicable, Wardveil requirements, and future Identity authority.

## Remaining Milestone 1 work

Milestone 1 remains open. Required work includes:

- embedded and approved sidecar metadata extraction;
- album/artist artwork discovery, storage policy, and reconciliation;
- recording, release, source-item, and playable-asset ingestion from verified scanned media;
- codec/container probing beyond filename-extension discovery;
- filesystem watchers or another approved event-driven change-detection mechanism where appropriate;
- Recently Added state and queries;
- higher-level Favorites, Ratings, and Recently Played service/API integration;
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

Those results establish only the checks executed against the merged Development source. They do not establish production deployment, runtime Platform-System acceptance, backup/restore acceptance, current-Stable Glaze UI acceptance, release eligibility, or Stable status.

Any later source change requires validation appropriate to that new exact revision.
