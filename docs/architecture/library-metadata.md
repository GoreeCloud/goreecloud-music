# GoreeCloud Music — Embedded Metadata Extraction Foundation

**Lifecycle:** Milestone 1 Development architecture  
**Scope:** Read-only embedded tag extraction and durable per-file metadata state  
**Implementation status:** Candidate until merged to `main` and verified by required checks

## Purpose

This slice advances the verified source-preserving scanner by extracting a bounded set of embedded textual metadata from an exact scanner-observed file. It intentionally stops before canonical Recording, Release, Source Item, or Playable Asset ingestion so uncertain or incomplete tags cannot silently become canonical music identity.

## Initial format support

The initial extractor supports:

- MP3 files with ID3v2.3 or ID3v2.4 text frames;
- FLAC files with Vorbis Comment metadata blocks.

MP3 files without an ID3v2 tag produce an explicit `none` tag format rather than invented values. Scanner-supported containers not listed above remain unsupported by this extraction slice and fail explicitly if extraction is requested.

The normalized initial fields are:

- title;
- artist;
- album;
- album artist;
- genre;
- date/year text;
- track-number text;
- disc-number text.

Track/disc/date values remain textual facts at this layer. Parsing them into canonical structured identity belongs to later ingestion validation.

## Durable schema

Schema version 4 adds `library_file_metadata` keyed by the scanner `file_id`. Each record stores:

- owning library identity;
- normalized embedded tag fields;
- detected tag format;
- source byte size and modification timestamp copied from the scanner observation used for extraction;
- extraction timestamp.

The database continues to store metadata and references only. Original media bytes remain in the separately managed library filesystem.

## Freshness contract

Metadata extraction is permitted only when the currently opened regular file still matches the scanner observation's byte size and nanosecond modification timestamp. If the source has changed, extraction fails with a rescan-required condition instead of binding tags to stale file facts.

A stored metadata row is considered current only when its extraction snapshot still matches the latest `library_files` observation and that observation is not tombstoned as missing.

The size/mtime pair is a freshness boundary, not a content-integrity hash and not canonical Recording identity.

## Filesystem safety

Before extraction:

- the configured library root must still be an absolute, non-symlink directory;
- the stored relative path must remain inside that root;
- every current path component is checked with `Lstat` and symbolic-link components are rejected;
- the final path must open as a regular file;
- scanner size/mtime facts are verified on the opened file descriptor.

The extractor performs reads only and never rewrites tags, artwork, media payloads, filenames, or directory structure.

## Authorization

- Reading extracted metadata requires library `read` permission.
- Extracting/replacing durable metadata requires library `edit` permission.
- Durable library permission remains an application authorization fact, not a substitute for future GoreeCloud Identity, Privacy Shield, Wardveil Security, purpose, or session checks at service/API boundaries.

## Parsing limits and failure behavior

The extractor uses bounded tag/block sizes and rejects malformed sizes, malformed UTF encodings, invalid FLAC signatures, excessive comment counts, and unsupported containers. It skips compressed/encrypted ID3 frames rather than interpreting bytes it cannot safely decode.

No parser failure creates or updates a canonical Recording/Release identity. The caller receives an error and existing scanner/media state remains intact.

## Explicitly deferred

This slice does **not** implement:

- ID3v2.2;
- ID3v1 fallback;
- MP4/M4A/ALAC atoms;
- Ogg/Opus/Vorbis comments outside FLAC;
- WAV/AIFF metadata chunks;
- embedded or sidecar artwork extraction;
- lyrics extraction;
- audio fingerprinting or media hashing;
- duration/codec/bitrate/bit-depth/sample-rate probing;
- canonical Recording or Release matching/creation;
- Source Item or Playable Asset ingestion;
- bulk/background metadata refresh orchestration;
- filesystem event watchers;
- public library/search APIs or user-facing metadata UI;
- production Platform-System acceptance.

Those remain active Milestone 1 or later obligations and must not be inferred from this metadata extraction foundation.
