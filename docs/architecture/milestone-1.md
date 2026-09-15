# Milestone 1 — Native Multi-User Library

## Status

Development implementation is active on `main`. PR #7 merged the bounded SQLite persistence and per-user library-authorization foundation as `564ac6a792070996dc39b103232c39b4fca95074`. PR #9 merged the bounded source-preserving library-scanning and filesystem-reconciliation foundation as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`. PR #12 merged the bounded authorization-scoped Favorites, Ratings, and Recently Played persistence foundation as `168d19d07e9b32e0089a512b9e6b7964db76ece1`. PR #14 merged the bounded authorization-scoped Recently Added query foundation as `c20acc1a92ee71ee4173468d97a5980ccf396013`. PR #16 merged the bounded embedded metadata extraction/persistence foundation as `bc90686a18338afef650b73357f454d1be541afb`. PR #18 merged bounded explicit-identity canonical materialization as `f365047f33f732afb1e274cd1d4b4b401271c32a`. PR #19 then merged bounded source-verified MP3/FLAC codec-container probing as current authoritative `main` `5e59466da23e5800f13f19e377a9e852631276ef`.

PR #9's exact candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed Music CI `35019904248` and Platform Contract `35019905704`. Its merged revision passed post-merge Music CI `35020139101` and Platform Contract `35020140030`.

PR #12's exact candidate `2721d2b7660763534396be4017ab9c18e47078da` passed Music CI `35025974258` and Platform Contract `35025974812`. The merged `main` revision passed post-merge Music CI `35026228206` and Platform Contract `35026228918`.

PR #14's exact candidate `8058489efbaafcfd95fbe1b6144a4eb6bc909c6d` passed Music CI `35028058553` and Platform Contract `35028059131`. The merged `main` revision passed post-merge Music CI `35028245616` and Platform Contract `35028246120`.

PR #16's exact candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed Music CI `35030658032` and Platform Contract `35030658657`. The merged revision `bc90686a18338afef650b73357f454d1be541afb` passed post-merge Music CI `35030857299` and Platform Contract `35030858110`.

PR #18's exact candidate `b46d2628a7fd86db79dc2a0be17945908166a8fd` passed Music CI `35032631291` and Platform Contract `35032632700`. The merged revision `f365047f33f732afb1e274cd1d4b4b401271c32a` passed post-merge Music CI `35032841324` and Platform Contract `35032841752`.

PR #19's exact candidate `d3eb0dcf31dce6b1802f0a1d979d628e15ee7a7f` passed Music CI `35034688612` and Platform Contract `35034689104`. The merged revision `5e59466da23e5800f13f19e377a9e852631276ef` passed post-merge Music CI `35034883077` and Platform Contract `35034883624`.

This record describes implemented Milestone 1 foundations and remaining work. It is not Milestone 1 completion, release evidence, production persistence acceptance, recovery acceptance, or Stable qualification.

## Objective

Build the first durable native GoreeCloud Music library service on top of the accepted Milestone 0 domain, routing, authorization-request, and engine-neutral storage contracts.

Milestone 1 must establish a multi-user library without turning the database into media storage, weakening user isolation, or allowing a storage engine, extracted tags, or filename extensions to become the Music domain authority.

## Current bounded implementation

Merged PR #7 provides SQLite application-state persistence, schema-v2 `library_memberships`, transactionally preserved schema-v1 owner authorization, profile/library persistence, explicit read/edit/owner permission evaluation, fail-closed owner mutation rules, absolute library-root validation, persisted library visibility, storage-aware `/healthz`, configurable Development database path, and migration/authorization/persistence tests.

Merged PR #9 adds schema-v3 `library_files`, supported-audio discovery for the approved extension set, symbolic-link avoidance, source-preserving scanning, validated relative paths, Added/Updated/Unchanged/Missing/Restored reconciliation, missing-file tombstones, permission-gated scan/file-state access, v2→v3 migration coverage, and source-preservation/isolation/unsafe-path tests.

Merged PR #12 adds profile-owned Favorite, Rating, and Recently Played operations; current library-read authorization checks; membership-filtered Favorite/Recently Played retrieval; independent profile state; validation bounds; and default-deny/grant/revocation/isolation tests. It deliberately adds no schema version because the relevant application-state entities already existed.

Merged PR #14 adds `RecentlyAddedForProfile` over canonical `recordings.added_at` state with newest-first deterministic ordering, current membership filtering, revoked-library suppression, and focused isolation/validation tests. It deliberately adds no schema version or dependency and does not implement Home UI or HTTP API exposure.

Merged PR #16 adds:

- forward-only schema migration from version 3 to version 4;
- durable `library_file_metadata` keyed by scanner `file_id`;
- bounded, read-only MP3 ID3v2.3/v2.4 extraction;
- bounded, read-only FLAC Vorbis Comment extraction;
- normalized title, artist, album, album artist, genre, date/year text, track-number text, and disc-number text;
- source size/mtime snapshot binding and `Current` evaluation against the latest scanner observation;
- explicit stale state after source changes or missing-file tombstones;
- library edit permission for extraction/persistence and read permission for retrieval;
- rooted source access using Go `os.Root`, explicit symbolic-link rejection, regular-file validation, and scanner snapshot verification before parsing and again before persistence;
- bounded parser limits and explicit malformed/unsupported-container failures;
- tests covering schema-v3→v4 migration, persistence, authorization, source changes, missing files, malformed metadata, unsupported containers, and symlink substitution.

Merged PR #18 then adds an explicit-identity materialization transaction from one current scanner/metadata observation into existing Recording, Release, GoreeCloud Server Source Item, and Playable Asset state. The caller must supply validated Recording and Release IDs. Source Item and Playable Asset identity are deterministic from the owning library and durable scanner `file_id`; edit/owner authorization and scanner/metadata freshness are rechecked; the rooted source snapshot is verified again before commit; exact retries are idempotent; and conflicting canonical/source/path bindings fail closed. This path does not infer recording equivalence from tag similarity.

Merged PR #19 adds bounded media format probing for an already-materialized playable asset. MP3 probing requires actual MPEG Layer III frame evidence, optionally after a validated ID3v2.3/v2.4 prefix; FLAC probing requires the native `fLaC` signature and valid first STREAMINFO block shape. The operation rechecks edit/owner authorization, scanner state, exact source/asset/path/size binding, rooted source safety, and source freshness, then persists only `codec` and `container`. Failed probes preserve prior format state. It does not establish duration, bitrate, sample rate, channels, bit depth, ReplayGain, content integrity, playback fitness, or identity equivalence.

## Data boundary

Original music files remain in GoreeCloud-controlled library storage. SQLite stores application state, file references/observations, normalized metadata facts, canonical identity/materialization records, and bounded format facts only. Media-file loss and database loss are separate recovery domains.

A `library_files` row is a filesystem observation, not a canonical Recording identity, content fingerprint, or playable-asset qualification.

A `library_file_metadata` row is a normalized extraction snapshot associated with one scanner observation. Its size/mtime binding is a freshness mechanism, not a content-integrity hash or canonical recording matcher.

PR #18 establishes canonical materialization only under explicit caller-supplied Recording/Release identity and verified current source facts. PR #19 verifies actual bytes for the bounded MP3/FLAC codec-container subset rather than trusting extensions, but those format facts are not a content hash, playback qualification, or equivalence decision.

Favorites, ratings, ordinary listening history, and Recently Added views are application state derived from authorized library records. They do not replace operational telemetry, security/audit records, or Privacy Shield purpose/retention controls where those controls apply.

Reusable secrets and external-provider credentials are outside ordinary Music application-state records.

## Authorization boundary

Library membership is explicit and fail-closed. A profile that has no membership receives no library permission. Creating a library creates only the owner's membership. Granting another profile access requires an explicit operation.

The schema-v1→v2 migration preserves existing library ownership by materializing owner memberships in the same transaction that creates `library_memberships`. The schema-v2→v3 migration adds file observations without altering membership authority. The schema-v3→v4 migration adds metadata state without changing library authorization facts or source media.

Reading scanner state requires library read permission. Running a scan or directly reconciling observations requires library edit permission. Extracting/replacing durable metadata requires library edit permission; reading stored metadata requires library read permission. Explicit canonical materialization and media-format probing require current edit/owner authority and independently revalidate their source bindings/freshness.

Favorites, ratings, Recently Played, and Recently Added retrieval continue to re-evaluate current library membership at the applicable operation boundary.

The backend provides durable authorization facts; it does not replace higher-level authorization. Service/API operations must continue to validate the requesting profile, action, purpose, Privacy Shield authorization where applicable, Wardveil requirements, and future Identity authority.

## Remaining Milestone 1 work

Milestone 1 remains open. Required work includes:

- approved sidecar metadata extraction;
- embedded metadata/container formats beyond the current MP3 ID3v2.3/v2.4 and FLAC Vorbis Comment subset;
- album/artist artwork discovery, storage policy, and reconciliation;
- automatic Recording/Release matching and duplicate-equivalence policy without silent substitution;
- broader/background canonical ingestion and reconciliation beyond the explicit-identity transaction;
- richer media probing, including duration, bitrate, sample rate, channel count, bit depth, ReplayGain, content hashes/integrity, and additional planned audio formats;
- filesystem watchers or another approved event-driven change-detection mechanism where appropriate;
- Home/user-facing Recently Added integration and API exposure;
- higher-level Favorites, Ratings, Recently Played, metadata, ingestion, and format service/API integration;
- complete multi-user isolation tests across every remaining library query/mutation path;
- library and local search APIs;
- migration/export and recovery controls;
- Everkeep-aligned backup/restore design and tested restoration;
- production authentication/session integration;
- observability, performance, corruption/failure, and concurrency acceptance.

## Verification boundary

PR #7's merged revision `564ac6a792070996dc39b103232c39b4fca95074` passed post-merge Music CI `35017128457` and Platform Contract `35017129147`.

PR #9's exact source candidate `a6579a7e029881ef80d8d202833de19deceb2da4` passed Music CI `35019904248` and Platform Contract `35019905704`; merged revision `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac` passed post-merge Music CI `35020139101` and Platform Contract `35020140030`.

PR #12's exact source candidate `2721d2b7660763534396be4017ab9c18e47078da` passed Music CI `35025974258` and Platform Contract `35025974812`; merged revision `168d19d07e9b32e0089a512b9e6b7964db76ece1` passed post-merge Music CI `35026228206` and Platform Contract `35026228918`.

PR #14's exact source candidate `8058489efbaafcfd95fbe1b6144a4eb6bc909c6d` passed Music CI `35028058553` and Platform Contract `35028059131`; merged revision `c20acc1a92ee71ee4173468d97a5980ccf396013` passed post-merge Music CI `35028245616` and Platform Contract `35028246120`.

PR #16's exact source candidate `fc9613d94493e2d5c355f3aba7b6bd2a24a2d334` passed Music CI `35030658032`, including exact-source verification, formatting, vet, race tests, and service build, and Platform Contract `35030658657`. It merged as `bc90686a18338afef650b73357f454d1be541afb`, which passed post-merge Music CI `35030857299` and Platform Contract `35030858110`.

PR #18's exact source candidate `b46d2628a7fd86db79dc2a0be17945908166a8fd` passed Music CI `35032631291` and Platform Contract `35032632700`; merged revision `f365047f33f732afb1e274cd1d4b4b401271c32a` passed post-merge Music CI `35032841324` and Platform Contract `35032841752`.

PR #19's exact source candidate `d3eb0dcf31dce6b1802f0a1d979d628e15ee7a7f` passed Music CI `35034688612` and Platform Contract `35034689104`; merged revision `5e59466da23e5800f13f19e377a9e852631276ef` passed post-merge Music CI `35034883077` and Platform Contract `35034883624`.

Those results establish only the checks executed against the merged Development source. They do not establish production deployment, runtime Platform-System acceptance, backup/restore acceptance, current-Stable Glaze UI acceptance, release eligibility, or Stable status.

Any later source change requires validation appropriate to that new exact revision.
