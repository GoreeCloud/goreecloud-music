# GoreeCloud Music — Planned Features and Capabilities

> **Repository mirror only.** The authoritative product record is `GoreeCloud/Projects/Project Specification — Music.docx`. This Markdown file exists for source-controlled continuity and must remain materially synchronized with the authoritative Drive specification. Planned or proposed capability text is not implementation evidence.

**Status:** Active Development — planned capability catalog established; provider-specific expansion remains proposed unless separately approved and verified  
**Application:** GoreeCloud Music  
**Capability identity:** GoreeCloud Resonance  
**Canonical repository:** `GoreeCloud/goreecloud-music`  
**Design system:** latest approved Stable Glaze UI release  
**Security:** Wardveil Security  
**Privacy:** Privacy Shield  
**Continuity:** Everkeep

## Current reconciliation — September 15, 2026

This specification mirror records the same product direction as the authoritative Drive specification. The canonical repository was newly initialized on September 15, 2026 and did not provide implementation evidence sufficient to treat the features below as implemented. The YouTube source integration remains a proposed optional expansion subject to technical, provider-policy, licensing, privacy, security, authorization, and legal review.

## Product principles

- Original GoreeCloud-owned native application; Navidrome/OpenSubsonic are references and interoperability targets, not the permanent complete-application foundation.
- Multi-user, self-hosted-first, privacy-by-default, explicit authorization, and source-transparent.
- Source and availability are separate first-class metadata.
- No silent recording substitution across providers.
- Current-Stable Glaze UI, Wardveil Security, Privacy Shield, and Everkeep remain mandatory platform contracts where applicable.
- Roadmap/specification presence never proves implementation or Stable status.

## Planned capability model

### Unified sources and state

- Unified private GoreeCloud Server library plus optional approved online discovery/playback.
- Explicit source labels such as **GoreeCloud Server**, proposed **YouTube**, and **Internet Radio**.
- Separate availability states: **Online**, **Offline**, **Downloaded**, **Cached**, **Unavailable**.
- Source identity travels with queue items, playlist entries, history, downloads, and handoff.
- Compact Glaze UI badges across Now Playing, mini-player, queue, search, menus, playlists, Credits, and downloads.
- Accessibility does not rely on color alone.

### Smart playback routing

- Prefer exact authorized downloaded copy, then authorized server copy, then approved online source according to user preference.
- Preferences: Prefer Local / Offline, Prefer GoreeCloud Server, Ask When Multiple Sources Exist.
- Match artist, album, duration, identifiers, release/edition metadata, and other signals; never silently replace a recording by title similarity alone.

### Search and library

- Search tracks, artists, albums, album artists, playlists, genres, composers, lyrics, years, and other supported metadata.
- Combined or scoped search for All, Server, proposed YouTube, Downloaded, and Offline.
- Self-hosted multiple libraries with per-user authorization, personal/shared scopes, incremental scans, filesystem change detection, embedded/sidecar metadata, artwork, recent activity, favorites, ratings, and persistent queues.

### Audio and player

- Planned codecs: FLAC, ALAC, WAV, AIFF, AAC, MP3, Opus, Ogg Vorbis.
- Original-quality/lossless playback where supported, HTTP ranges, gapless, ReplayGain/equivalent normalization, crossfade, configurable buffering, adaptive transcoding.
- Quality profiles: Efficient, Balanced, Lossless, Original Quality, optionally conditioned on Wi-Fi/Ethernet/cellular/roaming.
- Expose codec, bitrate, bit depth, sample rate, channels, transcode state, and original-vs-transcoded status where known.
- Advanced Now Playing, mini-player, and desktop floating/capsule player.

### Discovery, recommendations, and radio

- Home, Trending, Quick Picks, new releases, charts, genres, moods, decades, related music, radios, mixes.
- Made For You, Recently Played, Continue Listening, Heavy Rotation, Forgotten Favorites, Recently Added, Jump Back In, Daily Mixes, Family Favorites where authorized, recommended albums/artists, Discovery Mixes.
- Explore by genre, mood, decade, release date, high-resolution audio, soundtrack, live release, compilation, and other useful categories.
- Local-first recommendation engine using permitted user signals; optional external recommendation providers.
- Future local audio analysis may evaluate tempo, key, loudness, energy, and acoustic similarity.
- Song, Artist, Album, Playlist, Genre, Mood, Library, and authorized Family Radio with continuously extended queues and future tuning controls.

### Playlists and offline

- Personal, shared, collaborative, family, smart, imported/exported, and optional public-link playlists.
- Private, Shared View, Collaborative, Family, Public Link, expiring/revocable sharing, and collaborator attribution.
- Smart playlist rules for metadata, ratings/favorites, play/skip counts, recency, quality, source, and download state.
- Authorized downloads of songs, albums, playlists, artists, and smart playlists.
- Smart Downloads, storage quotas, source-aware storage manager, network/charging policies where supported.
- External-provider offline copies only when technically supported, provider-compliant, authorized/licensed, and lawful.

### Lyrics, metadata, analytics, and collaboration

- Embedded, `.lrc`, sidecar, synchronized, translated/transliterated, pronunciation-assisted, full-screen, karaoke, and future word-level lyrics.
- Rich credits and technical metadata including contributors, label/copyright, ISRC, catalog number, dates, genre, BPM/key, audio properties, source, and availability.
- Private per-user history/analytics and monthly/annual GoreeCloud Music Recap.
- Listen Together with invitations, synchronized playback, queue roles, participant visibility, and explicit private-track authorization.
- Independent user profiles and family experiences that never make private listening histories administratively visible by default.

### Interoperability and API

- Internet Radio with truthful stream semantics.
- Folder scan, direct upload where appropriate, M3U import/export, application-state export, Navidrome/Subsonic-compatible migration, potential Jellyfin music migration, future Apple/iTunes metadata migration.
- Commercial playlist migration is metadata migration, not authorization to copy catalog audio.
- Native GoreeCloud Music API under an application-owned namespace such as `/api/v1/`.
- OpenSubsonic compatibility as a subordinate compatibility layer.

### Clients and output

- Responsive web application and PWA where appropriate.
- Native Linux delivery.
- Native Android application.
- Planned native iOS application.
- Android Auto and planned CarPlay.
- Android TV / Google TV with possible later Apple TV, Roku, Chromecast-compatible, AirPlay-compatible, DLNA/UPnP support.
- Device transfer, per-device settings/quality, and future multi-room playback where synchronization is supported.
- Docker and Podman/OCI delivery for the self-hosted server where containerization applies.

### External integrations, privacy, and security

- Optional scrobbling, Discord Rich Presence, artwork, metadata, lyrics, discovery, and recommendation providers.
- No ad identifiers, no required third-party tracking, minimized telemetry, private listening data, and operational monitoring separated from listening history.
- Wardveil Security for authentication, authorization, isolation, admin boundaries, sharing/download authorization, abuse protection, secret handling, and safe failure.
- GoreeCloud Identity, Search, Notify, Manager, Everkeep, Wardveil Security, Privacy Shield, and Glaze UI integrate through their authoritative contracts.

## Recommended track identity model

`Source` answers **where is this music coming from?**  
`Availability` answers **can I play this without reaching that source right now?**

Examples:

- `Song Title — Artist` → `GoreeCloud Server · Offline · FLAC 24-bit/96 kHz`
- `Song Title — Artist` → `YouTube · Online · Opus`

This two-dimensional model lets future sources be added without redefining Online/Offline semantics.

## YouTube proposal boundary

YouTube remains optional and proposed. Its implementation requires a documented provider adapter, approved playback mechanism, privacy/data-flow analysis, authentication model if any, rate/failure behavior, replacement path, and provider-policy/licensing/legal review. `YouTube · Downloaded` is valid only when a separately approved method permits that specific offline copy. The self-hosted GoreeCloud library must remain operational when YouTube is disabled or unavailable.

## Verification boundary

This file is documentation, not proof of implementation. Features must not be represented as implemented, complete, production-ready, or Stable without applicable authoritative source, test, review, release, runtime, security, privacy, recovery, and current-Stable Glaze UI evidence.
