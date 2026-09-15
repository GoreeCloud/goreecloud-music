# GoreeCloud Music — Features

This file summarizes the planned product capability families defined by the authoritative GoreeCloud Music specification and `FEATURE-ROADMAP.md`. Unless a capability has separate verified implementation evidence, it is **planned**, not implemented.

## Core library and playback

- Self-hosted multi-user music libraries with explicit authorization.
- Incremental scanning, metadata and artwork management, favorites, ratings, and persistent queues.
- FLAC, ALAC, WAV, AIFF, AAC, MP3, Opus, and Ogg Vorbis planning.
- Original-quality and lossless playback where supported, range streaming, gapless playback, normalization, crossfade, buffering, and adaptive transcoding.
- Efficient, Balanced, Lossless, and Original Quality profiles with connection-aware policies.
- Deterministic playback routing with explicit route reasons and no silent recording substitution.

## Source and availability

- Independent source identity and availability state.
- GoreeCloud Server and Internet Radio as first-party/explicit source concepts.
- Optional provider adapters only through approved contracts.
- Online, Offline, Downloaded, Cached, and Unavailable availability states.
- Persistent source identity across queues, playlists, downloads, history, handoff, and resume.

## User experience

- Advanced Now Playing and persistent mini-player.
- Current-Stable Glaze UI across GoreeCloud-controlled graphical surfaces.
- Unified search with source-aware results and privacy-safe query routing.
- Home, Explore, recommendations, radio, rich metadata and Credits.
- Lyrics, synchronized lyrics, translations/transliteration, and planned karaoke experiences.
- Download and storage management with integrity and authorization state.

## Organization and sharing

- Personal, shared, collaborative, family, smart, imported/exported, and optional public-link playlists.
- Permission-aware sharing that never grants media access implicitly.
- Listen Together with explicit participant authorization.
- Multi-device synchronization and playback continuity.

## Privacy and intelligence

- Local-first recommendation processing.
- Private per-user listening history and analytics.
- Private Listening / Private Session planning.
- Monthly and annual GoreeCloud Music Recap experiences.
- No required behavioral advertising or third-party tracking.

## Clients and integrations

- Planned web/PWA, Linux, Android, iOS, Android Auto, CarPlay, television, output-transfer, and supported casting/receiver experiences.
- OpenSubsonic interoperability for authorized libraries while GoreeCloud Music’s native API remains authoritative.
- Optional scrobbling, rich-presence, artwork, metadata, lyrics, discovery, and recommendation providers with independent controls.
- Optional external providers must degrade independently so local browsing and authorized offline playback remain usable.

## Platform integration

GoreeCloud Music must evaluate and integrate the current applicable contracts for GoreeCloud Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze UI, GoreeCloud Mesh, and GoreeCloud Identity. No platform-system declaration is proof of runtime acceptance by itself.
