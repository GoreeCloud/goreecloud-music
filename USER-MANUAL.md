# GoreeCloud Music — Development Preview User Manual

## Status

GoreeCloud Music is in Active Development. Milestone 0 is an architecture foundation and is not a usable music application or Stable release.

## Running the development service

Requirements: Go 1.27.1 or a compatible supported Go 1.27 release.

```bash
go run ./cmd/musicd
```

The service binds to `127.0.0.1:8080` by default. A different development bind may be supplied through `GOREECLOUD_MUSIC_ADDR`.

## Available checks

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/api/v1/system/info
```

These endpoints only confirm the development process and build identity. They do not provide music-library or playback functionality.

## Not available yet

Library import/scanning, authentication, music playback, playlists, downloads, recommendations, online providers, web/Android/Linux clients, Glaze UI, and production deployment are not implemented in Milestone 0.
