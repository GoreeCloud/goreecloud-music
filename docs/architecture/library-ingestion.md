# GoreeCloud Music — Explicit Library-File Ingestion Foundation

**Lifecycle:** Milestone 1 Development architecture  
**Scope:** Atomic ingestion from one verified scanner/metadata observation into existing Recording, Release, Source Item, and Playable Asset storage  
**Implementation status:** Candidate until merged to `main` and verified by required checks

## Purpose

This slice establishes the transaction boundary between verified scanner/metadata state and the existing canonical Music storage entities. It deliberately does not invent a title/artist matching algorithm or treat extracted tags as sufficient proof that two files are the same recording.

## Identity boundary

The caller must supply validated canonical `RecordingID` and `ReleaseID` values. GoreeCloud Music does not derive those canonical identities from title, artist, album, filename, or other similarity heuristics in this slice.

The local GoreeCloud Server `SourceItemID` and `PlayableAssetID` are deterministically derived from the owning library and durable scanner `file_id`. This makes retries idempotent while keeping source-local identity separate from canonical recording/release identity.

A conflicting existing canonical ID, source item, or media-path binding causes the transaction to fail. Existing canonical records are not silently rewritten to fit new tag text.

## Preconditions

Ingestion requires all of the following:

- a valid profile and library identity;
- current library `edit` or `owner` permission re-evaluated inside the transaction;
- a non-missing `library_files` observation;
- a current `library_file_metadata` snapshot whose source size and modification timestamp still match the latest scanner observation;
- a source file that still passes the rooted/symlink-safe regular-file snapshot check;
- non-empty title and album metadata;
- caller-supplied valid Recording and Release IDs.

If any precondition fails, no partial canonical/source state is committed.

## Transaction behavior

A successful transaction materializes or safely reuses:

1. the explicit Release identity using the extracted album title and album-artist fallback;
2. the explicit Recording identity using extracted title/artist and that Release identity;
3. a deterministic GoreeCloud Server Source Item tied to the scanner `file_id`;
4. a deterministic Playable Asset tied to the library-relative source path and observed byte size.

Codec, container, duration, media hash, release-date normalization, and other fields that require later probing or validation remain unset rather than guessed from a filename extension or unparsed tag string.

Repeated ingestion of the same current file with the same explicit canonical identities is idempotent. Reuse of an existing identity with conflicting library/title/artist/release/path facts fails closed.

## Source preservation

The ingestion path does not modify, rename, move, tag, transcode, copy, or delete original media. Original files remain outside application database state. The file is opened read-only only to reconfirm the scanner snapshot before committing application state.

The current size/mtime freshness boundary is not a content-integrity hash. Media hashing and codec/container probing remain separate work.

## Explicitly deferred

This slice does **not** implement:

- automatic canonical Recording or Release matching;
- duplicate-recording equivalence decisions;
- approved sidecar metadata;
- artwork ingestion;
- embedded metadata formats beyond the current bounded extractor;
- codec/container/duration/bitrate/bit-depth/sample-rate probing;
- media content hashing or integrity qualification;
- bulk/background ingestion orchestration;
- event-driven filesystem watchers;
- public library/search APIs or user-facing ingestion UI;
- production Identity, Privacy Shield, Wardveil Security, Everkeep, deployment, release, or Stable acceptance.

Those remain active Milestone 1 or later obligations.
