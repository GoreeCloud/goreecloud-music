# GoreeCloud Music

GoreeCloud Music is a privacy-first, self-hosted, multi-user music library, streaming, discovery, radio, playlist, recommendation, and offline-listening service for GoreeCloud.

The project is being developed as original GoreeCloud-owned software. Its long-term architecture combines self-hosted ownership and interoperability with a polished modern listening experience while keeping source audio, metadata, listening history, recommendations, and user state under the operator's control.

**GoreeCloud Music — Powered by Resonance.**

## Current development state

The initial native foundation is merged to `main`. Milestone 1 development now includes:

- Go HTTP API service with bounded `/healthz` and `/api/v1/about` endpoints.
- React + TypeScript web-client foundation.
- Initial Glaze UI music application shell and persistent player surface.
- PostgreSQL development environment, runtime driver wiring, bounded connection lifecycle, and embedded schema migrations.
- Controlled migration execution behind `GOREECLOUD_MUSIC_AUTO_MIGRATE`, which is disabled by default.
- Multi-user domain types for users, libraries, memberships, artists, albums, tracks, and track files.
- Persistence interfaces plus an initial PostgreSQL store implementation.
- Request-principal and library authorization foundations with a development-only identity middleware that is disabled by default.
- Persistence-backed `GET /api/v1/me/libraries` support when a database and identity context are configured.
- Recursive audio-file discovery for common music formats and local ffprobe metadata extraction foundations.
- Domain, authorization, scanner, metadata, and HTTP API tests.
- Docker and CI foundations.

Production deployment is not approved. Production authentication, scanner reconciliation, metadata/artwork persistence ingestion, streaming, transcoding, OpenSubsonic compatibility, and native clients remain under development.

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
