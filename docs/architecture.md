# Architecture

## Product boundary

GoreeCloud Music is an original GoreeCloud-owned, privacy-first, self-hosted, multi-user music service. The native API and application model are authoritative. OpenSubsonic compatibility is planned as an interoperability layer rather than the internal architecture.

## Initial components

- **API service:** Go HTTP service.
- **Web client:** React and TypeScript with Glaze UI foundations.
- **Primary database:** PostgreSQL for application state and metadata.
- **Source audio:** Ordinary files on GoreeCloud-controlled storage; the database must not become the only usable representation of the library.
- **Transcoding:** FFmpeg is planned as a supporting dependency when transcoding is implemented.

## Trust boundaries

User identity, library membership, private listening history, playlists, recommendations, downloads, and administrative operations must be authorized independently. Administrative service management does not automatically grant permission to inspect private listening activity.

No reusable credential belongs in source control. External metadata, lyric, scrobbling, recommendation, or discovery providers must remain optional and explicitly configured.

## Planned API boundaries

- `/api/v1/` — first-party GoreeCloud Music API.
- `/rest/` — planned OpenSubsonic compatibility surface.
- `/healthz` — bounded operational health endpoint.

## Development order

1. Repository and architecture foundation.
2. Identity, authorization, PostgreSQL schema, libraries, scanner, and metadata extraction.
3. Browse/search, range streaming, queue, playlists, favorites, and ratings.
4. Original/lossless playback and adaptive transcoding.
5. Lyrics, advanced metadata, smart playlists, and credits.
6. Private recommendations and dynamic radio.
7. OpenSubsonic interoperability and migration tooling.
8. Collaborative playlists and offline APIs.
9. Native clients and production-hardening gates.
