# GoreeCloud Music — Library Scanning Foundation

**Lifecycle:** Milestone 1 Development architecture  
**Scope:** Supported-audio discovery and durable filesystem reconciliation  
**Implementation status:** Candidate until merged to `main` and verified by required checks

## Purpose

This slice establishes the filesystem-observation layer required before GoreeCloud Music can safely ingest metadata into canonical Recording, Release, Source Item, and Playable Asset entities.

The scanner does not treat database state as the source of truth for original media bytes. The configured library filesystem remains authoritative for the user's media files.

## Invariants

- Scanning never writes, renames, moves, deletes, or rewrites source media.
- Symbolic links are not followed during discovery, and a symbolic-link library root is rejected.
- Only regular files with explicitly supported audio extensions become observations.
- Stored file paths are library-relative and must not escape the configured root.
- A missing source file is tombstoned with `missing_since`; its observation identity is not silently deleted.
- A later reappearance clears the tombstone and is classified as restored.
- Read access is required to list library-file state.
- Edit access is required to scan or reconcile library-file state.
- The file-observation layer stores filesystem facts only; it does not claim tag parsing, artwork extraction, canonical recording identity, or playable-asset qualification.

## Supported discovery extensions

The initial scanner recognizes:

- `.flac`
- `.wav`
- `.aif`
- `.aiff`
- `.aac`
- `.m4a`
- `.mp3`
- `.opus`
- `.ogg`
- `.oga`

Extension recognition is case-insensitive. Extension recognition alone does not prove codec validity; format probing belongs to a later ingestion stage.

## Durable schema

Schema version 3 adds `library_files` with:

- deterministic `file_id`;
- owning `library_id`;
- normalized relative path;
- observed byte size;
- observed modification timestamp;
- first-seen and last-seen timestamps;
- optional missing-since tombstone.

The deterministic file identifier is scoped by library identity and relative path. It exists to preserve scan/reconciliation identity and is not a content hash or canonical Recording identity.

## Reconciliation states

A complete scan produces counts for:

- **Added** — a supported path not previously observed;
- **Updated** — an existing present path whose size or modification time changed;
- **Unchanged** — an existing present path with the same observed facts;
- **Missing** — a previously present path absent from the complete observation set;
- **Restored** — a tombstoned path observed again.

Repeated scans do not repeatedly count an already-tombstoned path as newly missing.

## Explicitly deferred

This slice does not implement:

- embedded or sidecar tag extraction;
- artwork discovery or storage;
- audio fingerprinting or content hashing;
- Recording/Release/Source Item/Playable Asset creation from scanned files;
- filesystem watchers or event-driven change notifications;
- library/search HTTP APIs;
- favorites, ratings, or recently-added presentation;
- production GoreeCloud Identity, Privacy Shield, Wardveil Security, or Everkeep runtime acceptance.

Those remain active Milestone 1 obligations and must not be inferred from this scanner foundation.
