# Architecture

## Product boundary

GoreeCloud Music is an original GoreeCloud-owned, privacy-first, self-hosted, multi-user music service. The native API and application model are authoritative. OpenSubsonic compatibility is planned as an interoperability layer rather than the internal architecture.

## Initial components

- **API service:** Go HTTP service with an explicit `internal/httpapi` boundary.
- **Web client:** React and TypeScript with Glaze UI foundations.
- **Primary database:** PostgreSQL for application state and metadata.
- **Source audio:** Ordinary files on GoreeCloud-controlled storage; the database must not become the only usable representation of the library.
- **Transcoding:** FFmpeg is planned as a supporting dependency when transcoding is implemented.

## Trust boundaries

User identity, library membership, private listening history, playlists, recommendations, downloads, and administrative operations must be authorized independently. Administrative service management does not automatically grant permission to inspect private listening activity.

No reusable credential belongs in source control. External metadata, lyric, scrobbling, recommendation, or discovery providers must remain optional and explicitly configured.

## Domain foundation

The initial domain model now establishes separate concepts for users, libraries, library memberships, artists, albums, tracks, and track files. Library membership is the first authorization boundary: a user can be granted read access independently from management access.

Source audio paths belong to track-file records rather than replacing music metadata. This keeps the application model capable of supporting multiple encodings or file variants for a logical track later without changing the track identity itself.

The first PostgreSQL migration mirrors these initial domain boundaries and uses foreign-key cleanup rules that preserve clear ownership relationships. It is development schema only; no production database migration has been authorized.

## Planned API boundaries

- `/api/v1/` — first-party GoreeCloud Music API.
- `/api/v1/about` — bounded development metadata for native clients and diagnostics.
- `/rest/` — planned OpenSubsonic compatibility surface.
- `/healthz` — bounded operational health endpoint.

## Development order

1. Repository and architecture foundation. **Implemented in Draft PR #1.**
2. Identity, authorization, PostgreSQL schema, libraries, scanner, and metadata extraction. **Domain/schema foundation in progress.**
3. Browse/search, range streaming, queue, playlists, favorites, and ratings.
4. Original/lossless playback and adaptive transcoding.
5. Lyrics, advanced metadata, smart playlists, and credits.
6. Private recommendations and dynamic radio.
7. OpenSubsonic interoperability and migration tooling.
8. Collaborative playlists and offline APIs.
9. Native clients and production-hardening gates.

## Current non-goals

The current branch does not yet implement password authentication, GoreeCloud Identity integration, persistent database access, filesystem scanning, metadata extraction, streaming, transcoding, or production deployment. These boundaries remain intentionally closed until their implementation and validation work is added.
