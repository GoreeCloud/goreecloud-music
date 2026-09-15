# Milestone 1 — Native Multi-User Library

## Status

Development implementation is active on `main`. PR #7 merged the bounded SQLite persistence and per-user library-authorization foundation as authoritative source commit `564ac6a792070996dc39b103232c39b4fca95074`. Post-merge Music CI run `35017128457` and Platform Contract run `35017129147` passed on that exact revision.

This record describes the implemented Milestone 1 foundation and remaining work. It is not Milestone 1 completion, release evidence, production persistence acceptance, recovery acceptance, or Stable qualification.

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

## Data boundary

Original music files remain in GoreeCloud-controlled library storage. SQLite stores application state and file references only. Media-file loss and database loss are therefore separate recovery domains.

Reusable secrets and external-provider credentials are outside ordinary Music application-state records.

## Authorization boundary

Library membership is explicit and fail-closed. A profile that has no membership receives no library permission. Creating a library creates only the owner's membership. Granting another profile access requires an explicit operation.

The schema-v1 → v2 migration preserves existing library ownership by materializing owner memberships in the same transaction that creates `library_memberships`, before the stored schema version advances.

The backend provides durable authorization facts; it does not replace higher-level authorization. Service/API operations must continue to validate the requesting profile, action, purpose, Privacy Shield authorization where applicable, Wardveil requirements, and future Identity authority.

## Remaining Milestone 1 work

Milestone 1 remains open. Required work includes:

- incremental scanning and filesystem-change detection;
- media-file identity/reconciliation that preserves source files;
- embedded and approved sidecar metadata extraction;
- album/artist artwork discovery, storage policy, and reconciliation;
- recording, release, source-item, and playable-asset ingestion;
- favorites and ratings service operations;
- recently-added and recently-played state;
- multi-user isolation tests across every library query/mutation path;
- library and local search APIs;
- migration/export and recovery controls;
- Everkeep-aligned backup/restore design and tested restoration;
- production authentication/session integration;
- observability, performance, corruption/failure, and concurrency acceptance.

## Verification boundary

The merged source revision `564ac6a792070996dc39b103232c39b4fca95074` passed post-merge Music CI run `35017128457`, including exact-source verification, formatting, vet, race tests, and service build. It also passed post-merge Platform Contract run `35017129147`.

Those results establish only the checks they execute against the merged Development source. They do not establish production deployment, runtime Platform-System acceptance, backup/restore acceptance, current-Stable Glaze UI acceptance, release eligibility, or Stable status.

Any later source change requires validation appropriate to that new exact revision.
