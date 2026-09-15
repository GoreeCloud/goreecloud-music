# Milestone 1 — Native Multi-User Library

## Status

Development implementation in Draft PR #7. This record describes the bounded Milestone 1 candidate and remaining work; it is not completion or release evidence.

## Objective

Build the first durable native GoreeCloud Music library service on top of the accepted Milestone 0 domain, routing, authorization-request, and engine-neutral storage contracts.

Milestone 1 must establish a multi-user library without turning the database into media storage, weakening user isolation, or allowing a storage engine to become the Music domain authority.

## Current bounded implementation

Draft PR #7 currently provides:

- SQLite application-state persistence through Go `database/sql` and `modernc.org/sqlite`;
- forward-only schema migration from version 1 to version 2;
- explicit `library_memberships` for per-profile authorization;
- profile and library persistence;
- atomic owner membership at library creation;
- read, edit, and owner permission evaluation;
- fail-closed owner mutation rules;
- absolute filesystem root validation for libraries;
- persisted library visibility across service/database reopen;
- storage-aware `/healthz` behavior;
- a configurable local Development database path;
- tests covering migration, persistence, default denial, explicit grant, owner protections, root-path validation, and SQLite safety settings.

## Data boundary

Original music files remain in GoreeCloud-controlled library storage. SQLite stores application state and file references only. Media-file loss and database loss are therefore separate recovery domains.

Reusable secrets and external-provider credentials are outside ordinary Music application-state records.

## Authorization boundary

Library membership is explicit and fail-closed. A profile that has no membership receives no library permission. Creating a library creates only the owner's membership. Granting another profile access requires an explicit operation.

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

Repository CI may prove that a specific source candidate formats, vets, passes race tests, and builds. Platform Contract validation may prove that its declaration satisfies the central Contract 0.2 validator. Neither result establishes production deployment, runtime Platform-System acceptance, backup/restore acceptance, current-Stable Glaze UI acceptance, release eligibility, or Stable status.

Any source change invalidates exact-head evidence for the previous commit and requires fresh validation.
