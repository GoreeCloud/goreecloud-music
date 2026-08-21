# GoreeCloud Music

GoreeCloud Music is a privacy-first, self-hosted, multi-user music library, streaming, discovery, radio, playlist, recommendation, and offline-listening service for GoreeCloud.

The project is being developed as original GoreeCloud-owned software. Its long-term architecture combines self-hosted ownership and interoperability with a polished modern listening experience while keeping source audio, metadata, listening history, recommendations, and user state under the operator's control.

## Current status

Initial native foundation in active development. Production deployment is not approved.

## Planned capabilities

- Independent multi-user accounts and private listening state.
- Multiple libraries with per-user access control.
- Original-quality and lossless playback.
- Adaptive transcoding when required.
- Playlists, smart playlists, collaborative playlists, favorites, and ratings.
- Local recommendations, mixes, and dynamic radio.
- Lyrics, rich metadata, artwork, and credits.
- Offline playback for first-party clients.
- Native API plus planned OpenSubsonic interoperability.
- Glaze UI, Wardveil Security, Privacy Shield, and Everkeep integration.

## Repository layout

```text
cmd/server/        Go API entry point
docs/              Architecture and project documentation
web/               React + TypeScript web client
.github/workflows/ Continuous integration
```

## Development

API:

```bash
go test ./...
go run ./cmd/server
```

Web client:

```bash
cd web
npm install
npm run dev
```

The API health endpoint is available at `GET /healthz` during development.

## Data ownership

Source audio remains ordinary files on GoreeCloud-controlled storage. Application metadata and state are stored separately so the music collection remains usable independently of GoreeCloud Music.

## License direction

The approved project specification selects GNU Affero General Public License v3.0 only (AGPL-3.0-only) as the preferred license. The canonical license file will be added before the first release candidate.
