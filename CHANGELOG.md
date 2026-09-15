---
title: "GoreeCloud Music — Repository Change Log"
document_type: "Repository Change Log"
status: "Active"
version: "v1.1"
classification: "Internal"
last_updated: "2026-09-15"
application: "GoreeCloud Music"
---

# GoreeCloud Music — Repository Change Log

This repository changelog records verified source and repository-documentation changes. The canonical GoreeCloud product changelog is `GoreeCloud/Changelogs/Change Log — Music.md`. Planned roadmap items are not completed changes merely because they appear in documentation.

## September 15, 2026 — Authorization-scoped Favorites, Ratings, and Recently Played foundation

- PR #12 merged as authoritative `main` commit `168d19d07e9b32e0089a512b9e6b7964db76ece1`.
- Exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed Music CI `35025974258` and Platform Contract `35025974812`.
- Post-merge `main` passed Music CI `35026228206` and Platform Contract `35026228918`.
- Added profile-owned Favorite set/clear and bounded listing operations using the existing `favorites` application-state entity.
- Added profile-owned 0–100 Rating set/read/clear operations using the existing `ratings` entity.
- Added profile-owned Recently Played event persistence and bounded recent-history listing using the existing `play_history` entity.
- Recording-scoped state operations require current library read authorization. Favorite/history retrieval re-evaluates membership so revoked library access is suppressed from normal Music surfaces.
- Tests cover default denial, explicit access grants, cross-profile isolation, independent rating state, authorization-revocation suppression, validation bounds, race testing, and service build.
- This change does not add a schema version or dependency and does not implement Recently Added, canonical media ingestion, product APIs, production authentication, recovery acceptance, release eligibility, or Stable status.

## September 15, 2026 — Markdown-first documentation alignment

- Repository documentation now points to the canonical Markdown product specification at `GoreeCloud/Projects/Project Specification — Music.md` and the Drive Markdown roadmap at `GoreeCloud/Feature Roadmap/GoreeCloud Music/FEATURE-ROADMAP.md`.
- `README.md`, `SPECIFICATIONS.md`, and `USER-MANUAL.md` are aligned with verified Milestone 1 persistence/authorization and scanner/reconciliation foundations.
- The duplicate repository product-specification mirror was retired so the governed Drive Markdown specification remains the single product-scope authority while `SPECIFICATIONS.md` remains repository-coupled implementation documentation.
- This documentation alignment does not change implementation state, complete Milestone 1, authorize release promotion, or establish Stable status.

## September 15, 2026 — Milestone 1 source-preserving scanner foundation

- PR #9 merged as authoritative `main` commit `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`.
- Exact candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed Music CI `35019904248` and Platform Contract `35019905704`.
- Post-merge `main` passed Music CI `35020139101` and Platform Contract `35020140030`.
- Schema version 3 adds durable `library_files` observations for source-preserving reconciliation.
- Supported-audio discovery avoids symbolic-link traversal and source-media mutation.
- Reconciliation distinguishes Added, Updated, Unchanged, Missing, and Restored observations; missing files are tombstoned rather than silently deleting identity.
- File-state listing requires read permission; scanning/reconciliation requires edit permission.
- Metadata/artwork extraction, canonical media ingestion, event-driven filesystem watching, library/search APIs, production recovery acceptance, and Milestone 1 completion remain open.

## September 15, 2026 — Scanner documentation reconciliation

- PR #10 reconciled `FEATURES.md`, `FEATURE-ROADMAP.md`, `NOTES.md`, and `docs/architecture/milestone-1.md` with the verified PR #9 implementation state.
- Exact documentation candidate `124b6f2c152878dcfbfa0778694ef37c2d5735a1` passed Music CI `35020559103` and Platform Contract `35020560112`.
- Merged documentation commit `4cdf5290a6fc71a95e759a5cf4f37b3bf9e816e0` passed post-merge Music CI `35021367454` and Platform Contract `35021370055`.

## September 15, 2026 — SQLite persistence and per-user library authorization foundation

- PR #7 merged as `564ac6a792070996dc39b103232c39b4fca95074`.
- Added SQLite Development application-state persistence, explicit `library_memberships`, profile/library persistence, owner protections, and schema-v1 owner-membership preservation during migration to schema v2.
- Post-merge Music CI `35017128457` and Platform Contract `35017129147` passed.
- Production persistence qualification, corruption/recovery evidence, and Everkeep-aligned recovery acceptance remain open.

## September 15, 2026 — Milestone 0 native architecture foundation

- PR #2 established the native Go architecture foundation, core domain identities, provider/authorization contracts, exact-match routing, queue identity, initial API contract, repository baseline, and CI foundation.
- PR #5 completed the bounded Milestone 0 storage/migration and Contract 0.2 reconciliation as `16a1a9b11cb8d6e06f7cf02fe7eedbb4725e7eda` after exact-head CI `34946346848` and Platform Contract `34946347663` passed.
- PR #6 synchronized the repository roadmap after the verified Milestone 0 merge.
