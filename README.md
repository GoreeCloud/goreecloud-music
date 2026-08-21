# GoreeCloud Music

GoreeCloud Music is a privacy-first, self-hosted, multi-user music library, streaming, discovery, radio, playlist, recommendation, and offline-listening service for GoreeCloud.

The project is being developed as original GoreeCloud-owned software. Its long-term architecture combines self-hosted ownership and interoperability with a polished modern listening experience while keeping source audio, metadata, listening history, recommendations, and user state under the operator's control.

## Current development state

Development is in the native foundation stage. Draft PR #1 now contains:

- Go HTTP API service with bounded `/healthz` and `/api/v1/about` endpoints.
- React + TypeScript web-client foundation.
- Initial Glaze UI music application shell and persistent player surface.
- PostgreSQL development environment and initial schema migration.
- Multi-user domain types for users, libraries, memberships, artists, albums, tracks, and track files.
- Persistence interfaces plus an initial PostgreSQL store implementation.
- Request-principal and library authorization foundations with a development-only identity middleware that is disabled by default.
- Recursive audio-file discovery for common music formats.
- Domain, authorization, scanner, and HTTP API tests.
- Docker and CI foundations, including a corrected container build that includes internal Go packages.

Production deployment is not approved. Production authentication, live database-driver wiring, runtime migration execution, metadata extraction, streaming, transcoding, OpenSubsonic compatibility, and native clients remain under development.

## Architecture direction

The first-party API is authoritative under `/api/v1/`. OpenSubsonic support is planned under `/rest/` as an interoperability layer rather than as the application's internal model.

Music originals remain ordinary files on controlled storage. Application-specific metadata and user state belong in PostgreSQL and must remain independently exportable and recoverable.

## Development

```bash
go test ./...
go run ./cmd/server
```

The development server listens on `:8080` unless `GOREECLOUD_MUSIC_ADDR` overrides it.

The web client is maintained under `web/`.

```bash
cd web
npm install --no-audit --no-fund
npm run dev
```

## Status

This repository is active development software and is not yet approved for production use.
