# GoreeCloud Music — Development Preview User Manual

## Status

GoreeCloud Music is in **Active Development** at **Milestone 1 — Native Multi-User Library**. It is not a complete music application or Stable release.

The repository contains verified backend foundations for native domain/storage architecture, SQLite application-state persistence, per-profile library authorization, source-preserving library-file discovery/reconciliation, profile-scoped Favorites/Ratings/Recently Played state, an authorization-scoped Recently Added query, and bounded embedded MP3/FLAC metadata extraction. These foundations are not yet exposed as a complete user-facing library-import, metadata, profile-state, Recently Added, search, or playback experience.

## Running the development service

Requirements: Go 1.27.1 or a compatible supported Go 1.27 release.

```bash
go run ./cmd/musicd
```

The service binds to `127.0.0.1:8080` by default. A different Development bind may be supplied through `GOREECLOUD_MUSIC_ADDR`.

The Development SQLite database path may be overridden with `GOREECLOUD_MUSIC_DB`.

## Available HTTP checks

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/api/v1/system/info
```

These endpoints confirm Development process/storage health and build identity. They do not provide a public music-library scan/import API, metadata API, profile-state API, Recently Added API, search API, or playback API.

## Implemented internal library foundations

Current merged Development source includes:

- persisted Music profiles and libraries;
- explicit read/edit/owner library memberships;
- forward schema migration through version 4;
- source-preserving discovery of supported regular audio files beneath an authorized library root;
- no symbolic-link traversal during discovery;
- no scanner mutation of source media;
- durable library-file observations;
- Added, Updated, Unchanged, Missing, and Restored reconciliation;
- missing-file tombstones;
- permission-gated file-state listing and reconciliation;
- profile-owned Favorites, Ratings, and Recently Played application-state operations for authorized recordings;
- current-membership filtering that suppresses Favorite/Recently Played state after library access is revoked;
- authorization-scoped Recently Added retrieval using canonical recording `added_at` state, with newest-first deterministic ordering and revoked-library suppression;
- durable `library_file_metadata` bound to exact scanner observations;
- read-only MP3 ID3v2.3/v2.4 and FLAC Vorbis Comment extraction for title, artist, album, album artist, genre, date/year text, track-number text, and disc-number text;
- library edit authorization for metadata extraction/storage and library read authorization for metadata retrieval;
- rooted source access, explicit symlink-component rejection, and scanner size/mtime verification before and after metadata parsing.

These are backend/internal capabilities. Their presence does not mean the application currently has a finished library browser, import workflow, metadata UI, Recently Added UI, profile-state UI, or playback surface.

## Not available yet

The following remain unimplemented or incomplete for user-facing use:

- approved sidecar metadata ingestion and metadata support for all planned media containers;
- artwork ingestion and reconciliation;
- canonical Recording/Release/Source Item/Playable Asset ingestion from scanned files;
- codec/container/duration/quality probing;
- background metadata refresh orchestration;
- event-driven filesystem watching;
- complete library/search/profile-state/metadata APIs and UI;
- Home/user-facing Recently Added integration;
- production authentication/session integration;
- music streaming, transcoding, and playback sessions;
- playlists and offline downloads as complete product features;
- recommendations and radio;
- online/external providers, including proposed YouTube support;
- first-class web, Android, Linux, iOS, automotive, or television clients;
- production deployment and recovery acceptance;
- Stable qualification.

See [`README.md`](README.md), [`SPECIFICATIONS.md`](SPECIFICATIONS.md), [`FEATURES.md`](FEATURES.md), and [`FEATURE-ROADMAP.md`](FEATURE-ROADMAP.md) for the verified Development boundaries and planned work.
