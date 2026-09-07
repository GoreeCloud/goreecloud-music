# GoreeCloud Music

GoreeCloud Music is a privacy-first, self-hosted, multi-user music library, streaming, discovery, radio, playlist, recommendation, and offline-listening service for GoreeCloud.

The project is being developed as original GoreeCloud-owned software. Its long-term architecture combines self-hosted ownership and interoperability with a polished modern listening experience while keeping source audio, metadata, listening history, recommendations, and user state under the operator's control.

**GoreeCloud Music — Powered by Resonance.**

## Current development state

The original native foundation is merged to `main`. Current first-party implementation includes:

- Go HTTP API service with bounded `/healthz` and `/api/v1/about` endpoints.
- React + TypeScript web-client foundation and initial Glaze UI music application shell.
- PostgreSQL runtime wiring, bounded connection lifecycle, and embedded schema migrations with controlled execution.
- Multi-user domain types, request-principal handling, persisted library membership, and separate library read/manage authorization boundaries.
- Authorization-aware Resonance Library scanning and PostgreSQL reconciliation for artists, albums, tracks, track files, metadata, and local album artwork.
- Authorized library browse APIs for albums and tracks plus authorized album-artwork delivery.
- Local ffprobe/FFmpeg-backed metadata and artwork handling without requiring an external metadata or artwork service.
- Authorized full and bounded single-range HTTP track streaming through the first-party `/api/v1/` surface.
- Resonance Player queue core with ordered queues, current-track state, Next/Previous navigation, Repeat Off/All/One, and in-development unplayed-tail shuffle behavior.
- Docker and CI foundations with Go and web validation.

Production deployment is not approved. Production GoreeCloud Identity integration, persistent per-user queues, native web playback wiring, shuffle-state persistence, favorites, ratings, playlists, transcoding, OpenSubsonic compatibility, offline clients, and broader native clients remain under development.

## Resonance

GoreeCloud Resonance is the approved capability identity for the first-party music feature platform. GoreeCloud Music remains the application and runtime authority; Resonance organizes its durable capability families without creating a separate service or permission model.

See `docs/resonance.md` and `docs/resonance.identity.json`.

## Architecture direction

The first-party API is authoritative under `/api/v1/`. OpenSubsonic support is planned under `/rest/` as an interoperability layer rather than as the application's internal model.

Music originals remain ordinary files on controlled storage. Application-specific metadata and user state belong in PostgreSQL and must remain independently exportable and recoverable.

## Development

```bash
go test ./...
go run ./cmd/server
```

The development server listens on `:8080` unless `GOREECLOUD_MUSIC_ADDR` overrides it.

`DATABASE_URL` enables PostgreSQL-backed runtime behavior. `GOREECLOUD_MUSIC_AUTO_MIGRATE=true` explicitly permits the server to apply embedded migrations at startup; the default is `false`. `GOREECLOUD_MUSIC_DEV_IDENTITY=true` enables the development-only request identity middleware and must never be treated as production authentication.

The web client is maintained under `web/`.

```bash
cd web
npm install --no-audit --no-fund
npm run dev
```

## Status

This repository is active development software and is not yet approved for production use.

## License

GoreeCloud Music is licensed under the GNU Affero General Public License, version 3 only (`AGPL-3.0-only`). See `LICENSE`.
