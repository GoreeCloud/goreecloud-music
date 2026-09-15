# GoreeCloud Music — Development Preview User Manual

## Status

GoreeCloud Music is in **Active Development** at **Milestone 1 — Native Multi-User Library**. It is not a complete music application or Stable release.

The repository contains verified backend foundations for native domain/storage architecture, SQLite application-state persistence, per-profile library authorization, and source-preserving library-file discovery/reconciliation. These foundations are not yet exposed as a complete user-facing library-import, search, or playback experience.

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

These endpoints confirm Development process/storage health and build identity. They do not provide a public music-library scan/import API, search API, or playback API.

## Implemented internal library foundations

Current merged Development source includes:

- persisted Music profiles and libraries;
- explicit read/edit/owner library memberships;
- forward schema migration through version 3;
- source-preserving discovery of supported regular audio files beneath an authorized library root;
- no symbolic-link traversal during discovery;
- no scanner mutation of source media;
- durable library-file observations;
- Added, Updated, Unchanged, Missing, and Restored reconciliation;
- missing-file tombstones;
- permission-gated file-state listing and reconciliation.

These are backend/internal capabilities. Their presence does not mean the application currently has a finished library browser, import workflow, metadata UI, or playback surface.

## Not available yet

The following remain unimplemented or incomplete for user-facing use:

- metadata/tag and artwork ingestion;
- canonical Recording/Release/Source Item/Playable Asset ingestion from scanned files;
- event-driven filesystem watching;
- complete library/search APIs and UI;
- favorites, ratings, and recent-activity service operations;
- production authentication/session integration;
- music streaming, transcoding, and playback sessions;
- playlists and offline downloads as complete product features;
- recommendations and radio;
- online/external providers, including proposed YouTube support;
- first-class web, Android, Linux, iOS, automotive, or television clients;
- production deployment and recovery acceptance;
- Stable qualification.

See [`README.md`](README.md), [`SPECIFICATIONS.md`](SPECIFICATIONS.md), [`FEATURES.md`](FEATURES.md), and [`FEATURE-ROADMAP.md`](FEATURE-ROADMAP.md) for the verified Development boundaries and planned work.
