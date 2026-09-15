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

## Storage and media boundary

Original music files remain in GoreeCloud-controlled library storage. SQLite stores application state and references/observations only; it does not embed original media bytes. Media-file loss and database-state loss remain separate recovery domains. Reusable authentication credentials and provider secrets remain outside ordinary Music application-state records.

The deterministic library-file identifier used by the scanner is a reconciliation identity derived from library identity and relative path. It is not a content hash and does not establish canonical Recording identity.

## Authorization boundary

Durable library membership supplies application authorization facts but does not replace higher-level GoreeCloud authority. API/service operations must continue to validate the requesting profile, operation, purpose, applicable Privacy Shield authorization, Wardveil requirements, and future GoreeCloud Identity/session authority.

Queue persistence or previously observed file state does not grant perpetual access. Authorization must be rechecked at the operation/playback boundary where required.

## Current status boundary

The verified foundations do **not** establish:

- embedded or sidecar metadata extraction;
- artwork discovery/processing;
- canonical Recording, Release, Source Item, or Playable Asset ingestion from scanned files;
- event-driven filesystem watchers/change notifications;
- favorites, ratings, or recent-activity service operations;
- library/search product APIs;
- actual audio streaming/transcoding/playback sessions;
- production authentication/session integration;
- real external provider adapters;
- production persistence, corruption, backup/restore, or Everkeep acceptance;
- user-facing web, Android, Linux, iOS, automotive, or television clients;
- current-Stable Glaze UI acceptance;
- release eligibility or Stable qualification.

Those remain later Milestone 1 or subsequent milestone obligations and must not be inferred from the persistence/scanner foundations.
