# GoreeCloud Music — Feature Roadmap

> Repository-side roadmap control. The corresponding Drive control is `GoreeCloud/Feature Roadmap/GoreeCloud Music/FEATURE-ROADMAP.docx`. The two records must remain materially synchronized with each other and with the authoritative `Project Specification — Music.docx`.

**Active roadmap control — As of September 15, 2026**

## Purpose

This roadmap records current planned and recommended GoreeCloud Music feature work without replacing the authoritative project record, repository implementation evidence, release gates, or GoreeCloud Tasks Management. Roadmap presence is not implementation evidence.

## Control records

| Field | Value |
|---|---|
| Application / Service | GoreeCloud Music |
| Authoritative project record | Project Specification — Music.docx |
| Canonical repository | `GoreeCloud/goreecloud-music` |
| Repository control | `FEATURE-ROADMAP.md` |
| Drive location | `GoreeCloud/Feature Roadmap/GoreeCloud Music/FEATURE-ROADMAP.docx` |

## Roadmap

| ID | Feature / obligation | Priority | Current state |
|---|---|---|---|
| FR-001 | Reconcile and maintain every current planned or recommended GoreeCloud Music feature against the authoritative project record, repository evidence, and applicable platform requirements. | High | Ongoing control |
| FR-002 | Move actionable feature obligations into GoreeCloud Tasks Management when required, avoiding duplicate task authority. | High | Ongoing control |
| FR-003 | Do not mark features implemented, complete, cancelled, or superseded without authoritative evidence and lifecycle reconciliation. | High | Ongoing control |
| FR-004 | Maintain GoreeCloud Music as an original GoreeCloud-owned native application using Navidrome/OpenSubsonic as references or compatibility targets rather than a permanent complete-application foundation. | High | Planned / governing architecture |
| FR-005 | Deliver a unified music experience combining authorized GoreeCloud Server libraries with optional approved online discovery/playback sources while preserving provider boundaries. | High | Planned; YouTube portion proposed |
| FR-006 | Implement first-class source identity metadata and compact Glaze UI source indicators for GoreeCloud Server, proposed YouTube, Internet Radio, and future explicitly supported providers. | High | Planned |
| FR-007 | Implement availability metadata independent from source, including Online, Offline, Downloaded, Cached, and Unavailable states. | High | Planned |
| FR-008 | Present source and availability metadata throughout Now Playing, mini-player, queue, search, menus, mixed playlists, Credits, and download management without relying on color alone. | High | Planned |
| FR-009 | Implement smart playback routing that prefers an exact downloaded copy, then authorized server copy, then approved online source according to user preference without silent recording substitution. | High | Planned |
| FR-010 | Implement unified search across private libraries and approved online sources with All, Server, YouTube, Downloaded, and Offline scopes where available. | High | Planned; YouTube scope proposed |
| FR-011 | Implement self-hosted multi-library music storage, authorization, incremental scanning, filesystem change detection, embedded/sidecar metadata, artwork, favorites, ratings, recently added/played, and persistent queues. | High | Planned |
| FR-012 | Support FLAC, ALAC, WAV, AIFF, AAC, MP3, Opus, and Ogg Vorbis with original-quality/lossless playback where supported, HTTP range streaming, gapless playback, normalization, crossfade, buffering, and adaptive transcoding. | High | Planned |
| FR-013 | Implement Efficient, Balanced, Lossless, and Original Quality profiles with connection-aware policies and codec/bitrate/bit-depth/sample-rate/channel/transcode diagnostics. | High | Planned |
| FR-014 | Build Advanced Now Playing as a signature Glaze UI experience with artwork, source, availability, quality, transport, queue, lyrics, Credits, device, volume, sleep timer, and contextual actions. | High | Planned |
| FR-015 | Implement persistent mini-player and desktop floating/capsule player with platform-appropriate always-on-top behavior and accessible Glaze UI motion/effects. | Medium | Planned |
| FR-016 | Implement optional online music discovery including recommendations, trending music, Quick Picks, new releases, charts, genres, moods, decades, related music, radios, and mixes with explicit source identity. | Medium | Planned |
| FR-017 | Implement Home experiences including Made For You, Recently Played, Continue Listening, Heavy Rotation, Forgotten Favorites, Recently Added, Jump Back In, Daily Mixes, Family Favorites when authorized, Recommended Albums/Artists, and Discovery Mixes. | Medium | Planned |
| FR-018 | Implement Explore browsing by genre, mood, decade, release date, high-resolution audio, soundtrack, live release, compilation, and other useful categories. | Medium | Planned |
| FR-019 | Implement a local-first private recommendation engine using permitted favorites, ratings, repeats, skips, playlists, artists, albums, genres, years, time, and sequence signals; keep external recommendation providers optional. | High | Planned |
| FR-020 | Evaluate future local audio-analysis signals such as tempo, key, loudness, energy, and acoustic similarity without making them a privacy-invasive or mandatory dependency. | Low | Future planned |
| FR-021 | Implement Song, Artist, Album, Playlist, Genre, Mood, Library, and authorized Family Radio with continuously extended queues and future tuning controls. | Medium | Planned |
| FR-022 | Implement personal, shared, collaborative, family, smart, imported/exported, and optional public-link playlists with search, multi-select, reorder, M3U support, migration, and mixed-source entries. | High | Planned |
| FR-023 | Implement playlist permissions for Private, Shared View, Collaborative, Family, Public Link, expiring/revocable links, and collaborative attribution without bypassing underlying library permissions. | High | Planned |
| FR-024 | Implement smart playlists based on artist, genre, release date, rating, favorites, play/skip count, recent activity, audio quality, source, and downloaded state. | Medium | Planned |
| FR-025 | Implement offline downloads for authorized songs, albums, playlists, artists, and smart playlists while treating user-owned server content as a normal offline capability. | High | Planned |
| FR-026 | Permit external-provider offline copies such as YouTube · Downloaded only when separately verified as technically supported, provider-compliant, licensed/authorized, and lawful. | High | Proposed; approval required |
| FR-027 | Implement Smart Downloads with user-defined storage targets, automatic selection/rotation, and clear manual-versus-automatic attribution. | Medium | Planned |
| FR-028 | Implement a Download and Storage Manager with usage breakdowns, deletion, offline quality, storage limits, Wi-Fi-only behavior, charging-only automation where supported, quotas, and source origin. | Medium | Planned |
| FR-029 | Implement advanced lyrics including embedded, .lrc, sidecar, line-sync, future word-level sync, transliteration, translation, pronunciation assistance, full-screen lyrics, karaoke, and permitted lyric-card sharing. | Medium | Planned |
| FR-030 | Implement rich metadata and Credits covering contributors, label/copyright, ISRC, catalog number, release dates, genre, BPM/key, audio diagnostics, lyrics, artwork, source, and availability. | Medium | Planned |
| FR-031 | Implement private per-user listening history and analytics including recently played, counts, listening time, top content, genre distribution, time-period views, listening-clock views, historical charts, and source breakdown. | Medium | Planned |
| FR-032 | Implement monthly and annual GoreeCloud Music Recap experiences with shareable Glaze UI recap cards and family-level aggregation only with explicit permission. | Low | Planned |
| FR-033 | Implement Listen Together with invitations, participant list, host/collaborative queue, synchronized playback, and explicit authorization for private tracks. | Medium | Planned |
| FR-034 | Implement independent multi-user profiles, favorites, ratings, history, recommendations, queues, downloads, settings, devices, playlists, smart playlists, and library permissions. | High | Planned / mandatory architecture |
| FR-035 | Implement Family Music with shared libraries, playlists, radio, optional shared favorites, and privacy boundaries that do not expose private histories to family administrators by default. | High | Planned |
| FR-036 | Implement user-configured Internet Radio with favorites, recent stations, station metadata/artwork when available, and truthful live-stream source semantics. | Medium | Planned |
| FR-037 | Implement import/export and migration for existing music folders, direct upload where appropriate, M3U, application-state export, Navidrome/Subsonic-compatible systems, potential Jellyfin music migration, and future metadata-only Apple/iTunes migration. | Medium | Planned |
| FR-038 | Keep commercial-streaming playlist migration metadata-only unless the user independently has authorization to obtain the underlying audio; migration does not authorize catalog copying. | High | Governing restriction |
| FR-039 | Provide OpenSubsonic compatibility for authorized libraries while keeping the native GoreeCloud Music API and internal architecture authoritative. | Medium | Planned |
| FR-040 | Implement a first-party /api/v1/ Music API for modern authentication/authorization, rich metadata, playback, offline sync, recommendations, radio, source/availability identity, and future features. | High | Planned |
| FR-041 | Deliver a responsive web application and installable PWA behavior where appropriate with keyboard/media-key support, notifications, playlist management, and adaptive Glaze UI. | High | Planned / mandatory delivery |
| FR-042 | Deliver a supported native Linux client/package with desktop integration, media keys, floating player, notifications, and documented release/rollback paths. | High | Planned / mandatory delivery |
| FR-043 | Deliver a first-class native Android application with background playback, media/lock-screen controls, offline downloads, sharing, widgets, dynamic Now Playing, and player/lyrics/queue/Credits gestures. | High | Planned / mandatory delivery |
| FR-044 | Plan a native iOS application that follows the same Music, privacy, security, source identity, and current-Stable Glaze UI contracts. | Medium | Planned |
| FR-045 | Implement Android Auto and plan CarPlay with voice-friendly browsing and driving-appropriate access to favorites, playlists, recent/downloaded music, radio, and recommendations. | Medium | Planned |
| FR-046 | Implement Android TV / Google TV and evaluate later Apple TV, Roku, Chromecast-compatible, AirPlay-compatible, and DLNA/UPnP playback ecosystems with large-screen lyrics and Now Playing. | Medium | Planned / future expansion |
| FR-047 | Implement device/output management for current device, playback transfer, device-specific settings/quality, and future multi-room playback where synchronization is supported. | Medium | Planned |
| FR-048 | Support individually controllable optional integrations such as scrobbling, Discord Rich Presence, artwork, metadata, lyrics, discovery, and recommendation providers without making them mandatory dependencies. | Medium | Planned |
| FR-049 | Enforce privacy-by-default behavior: no advertising identifiers, no required third-party tracking, minimized telemetry, provider controls, private listening data, operational/history separation, and local recommendations by default. | High | Planned / mandatory platform contract |
| FR-050 | Enforce Wardveil Security requirements for authentication, sessions, authorization, user isolation, administration, sharing, downloads, abuse controls, secure secret handling, safe failure, and release evidence. | High | Planned / mandatory platform contract |
| FR-051 | Integrate GoreeCloud Identity, Search, Notify, Manager, Everkeep, Wardveil Security, Privacy Shield, and Glaze UI through their authoritative contracts without bypassing Music authorization or privacy boundaries. | High | Planned / platform integration |
| FR-052 | Provide supported Docker and Podman/OCI delivery for the self-hosted/server component where containerization applies, with documented storage, backup, upgrade, rollback, monitoring, and recovery. | High | Planned / mandatory delivery |
| FR-053 | Keep source and availability as independent extensible track metadata so future providers can be added without redesigning the user-facing state model. | High | Planned / governing data model |
| FR-054 | Keep proposed YouTube integration optional and gated by technical feasibility, provider policy, licensing/authorization, privacy, security, legal review, failure behavior, and replacement-path documentation. | High | Proposed; approval required |
| FR-055 | Maintain current-Stable Glaze UI conformance across every GoreeCloud-controlled Music interface as a mandatory Stable release gate. | High | Ongoing release control |
| FR-056 | Maintain Everkeep-aligned backup/recovery and preservation for required application state while keeping original music files independent from application database state. | High | Planned / mandatory platform contract |

## Maintenance and synchronization

This roadmap and the Drive `FEATURE-ROADMAP.docx` must remain materially synchronized with each other and with the authoritative product specification. Update both copies whenever feature scope, priority, dependency, implementation status, cancellation, supersession, recommendation, or verification state materially changes.

No feature may be represented as complete or Stable solely because it appears here. Completion and lifecycle claims require applicable authoritative implementation, validation, review, release, production, security, privacy, recovery, and current-Stable Glaze UI evidence.

## Reconciliation rule

At each material feature change, reconcile this roadmap against the current authoritative project record, repository implementation state, applicable platform-system requirements, and GoreeCloud Tasks Management.
