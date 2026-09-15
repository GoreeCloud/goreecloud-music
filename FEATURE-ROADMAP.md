# GoreeCloud Music — Feature Roadmap

> Repository-side roadmap control. The corresponding Drive control is `GoreeCloud/Feature Roadmap/GoreeCloud Music/FEATURE-ROADMAP.docx`. The two records must remain materially synchronized with each other and with the authoritative `Project Specification — Music.docx`. Roadmap presence is not implementation evidence.

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
| FR-001 | Reconcile all planned/recommended Music features against the authoritative project record, repository evidence, and platform requirements. | High | Ongoing control |
| FR-002 | Move actionable feature obligations into GoreeCloud Tasks Management when required without creating duplicate task authority. | High | Ongoing control |
| FR-003 | Do not mark features implemented, complete, cancelled, or superseded without authoritative evidence and lifecycle reconciliation. | High | Ongoing control |
| FR-004 | Maintain Music as original GoreeCloud-owned native software; Navidrome/OpenSubsonic remain references/compatibility targets rather than a permanent complete-application foundation. | High | Planned / governing architecture |
| FR-005 | Deliver a unified experience combining authorized GoreeCloud Server libraries with optional approved online discovery/playback while preserving provider boundaries. | High | Planned; YouTube portion proposed |
| FR-006 | Implement first-class source identity metadata and compact Glaze UI source indicators for GoreeCloud Server, proposed YouTube, Internet Radio, and future supported sources. | High | Planned |
| FR-007 | Keep availability independent from source using Online, Offline, Downloaded, Cached, and Unavailable. | High | Planned |
| FR-008 | Present source and availability in Now Playing, mini-player, queue, search, menus, mixed playlists, Credits, and downloads without color-only semantics. | High | Planned |
| FR-009 | Route playback to an exact downloaded copy, authorized server copy, or approved online source according to policy without silent recording substitution. | High | Planned |
| FR-010 | Search private libraries and approved online sources with All, Server, YouTube, Downloaded, and Offline scopes where available. | High | Planned; YouTube scope proposed |
| FR-011 | Implement multi-library self-hosted storage, per-user authorization, scanning/change detection, metadata/artwork, favorites, ratings, recent activity, and persistent queues. | High | In progress — PR #7 merged and post-merge validated the SQLite application-state and per-user library-authorization foundation. PR #9 merged schema-v3 `library_files`, supported-audio discovery, source-preserving Added/Updated/Unchanged/Missing/Restored reconciliation, missing-file tombstones, no-symlink scanning, and permission-gated file-state access as `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac`; post-merge Music CI `35020139101` and Platform Contract `35020140030` passed. Metadata/artwork extraction, canonical media ingestion, event-driven change detection, favorites/ratings/recent activity, library/search APIs, production recovery acceptance, and remaining Milestone 1 work remain pending. |
| FR-012 | Support FLAC, ALAC, WAV, AIFF, AAC, MP3, Opus, and Ogg Vorbis with original/lossless playback where supported, ranges, gapless, normalization, crossfade, buffering, and adaptive transcoding. | High | Planned |
| FR-013 | Implement Efficient, Balanced, Lossless, and Original Quality profiles with connection-aware policies and playback diagnostics. | High | Planned |
| FR-014 | Build Advanced Now Playing with artwork, source, availability, quality, transport, queue, lyrics, Credits, device, volume, timer, and contextual actions. | High | Planned |
| FR-015 | Implement persistent mini-player and desktop floating/capsule player with accessible Glaze UI behavior. | Medium | Planned |
| FR-016 | Implement optional online discovery: recommendations, trending, Quick Picks, new releases, charts, genres, moods, decades, related music, radios, and mixes with explicit source identity. | Medium | Planned |
| FR-017 | Implement Home experiences including Made For You, Recently Played, Continue Listening, Heavy Rotation, Forgotten Favorites, Recently Added, Jump Back In, Daily Mixes, authorized Family Favorites, recommended albums/artists, and Discovery Mixes. | Medium | Planned |
| FR-018 | Implement Explore by genre, mood, decade, release date, high-resolution audio, soundtrack, live release, compilation, and other useful categories. | Medium | Planned |
| FR-019 | Implement local-first recommendations from permitted user signals while keeping external recommendation providers optional. | High | Planned |
| FR-020 | Evaluate future local audio-analysis signals such as tempo, key, loudness, energy, and acoustic similarity without creating a privacy-invasive dependency. | Low | Future planned |
| FR-021 | Implement Song, Artist, Album, Playlist, Genre, Mood, Library, and authorized Family Radio with continuously extended queues and future tuning controls. | Medium | Planned |
| FR-022 | Implement personal, shared, collaborative, family, smart, imported/exported, and optional public-link playlists with search, multi-select, reorder, M3U, migration, and mixed-source entries. | High | Planned |
| FR-023 | Implement Private, Shared View, Collaborative, Family, Public Link, expiring/revocable sharing, and collaborator attribution without bypassing library permissions. | High | Planned |
| FR-024 | Implement smart playlists based on metadata, ratings/favorites, play/skip counts, recency, quality, source, and downloaded state. | Medium | Planned |
| FR-025 | Implement authorized offline downloads for songs, albums, playlists, artists, and smart playlists; user-owned server content remains a normal offline capability. | High | Planned |
| FR-026 | Permit external-provider offline copies such as `YouTube · Downloaded` only when separately verified as technically supported, provider-compliant, licensed/authorized, and lawful. | High | Proposed; approval required |
| FR-027 | Implement Smart Downloads with storage targets, automatic selection/rotation, and manual-versus-automatic attribution. | Medium | Planned |
| FR-028 | Implement Download and Storage Manager with usage breakdowns, deletion, offline quality, limits, network/charging policies, quotas, and source origin. | Medium | Planned |
| FR-029 | Implement embedded, `.lrc`, sidecar, synchronized, translated/transliterated, pronunciation-assisted, full-screen, karaoke, and future word-level lyrics. | Medium | Planned |
| FR-030 | Implement rich metadata/Credits including contributors, label/copyright, ISRC, catalog number, dates, genre, BPM/key, audio diagnostics, lyrics, artwork, source, and availability. | Medium | Planned |
| FR-031 | Implement private per-user history/analytics including recent plays, counts, listening time, top content, genre distribution, time views, charts, and source breakdown. | Medium | Planned |
| FR-032 | Implement monthly/annual GoreeCloud Music Recap with shareable Glaze UI cards and family aggregation only with explicit permission. | Low | Planned |
| FR-033 | Implement Listen Together with invitations, participant list, host/collaborative queue, synchronized playback, and explicit authorization for private tracks. | Medium | Planned |
| FR-034 | Implement independent multi-user profiles, favorites, ratings, history, recommendations, queues, downloads, settings, devices, playlists, smart playlists, and library permissions. | High | Planned / mandatory architecture |
| FR-035 | Implement Family Music with shared libraries/playlists/radio, optional shared favorites, and privacy boundaries that do not expose private histories by default. | High | Planned |
| FR-036 | Implement user-configured Internet Radio with favorites, recent stations, metadata/artwork when available, and truthful live-stream semantics. | Medium | Planned |
| FR-037 | Implement folder scanning, appropriate direct upload, M3U, state export, Navidrome/Subsonic-compatible migration, potential Jellyfin music migration, and future Apple/iTunes metadata migration. | Medium | Planned |
| FR-038 | Keep commercial-streaming playlist migration metadata-only unless the user independently has authorization to obtain underlying audio. | High | Governing restriction |
| FR-039 | Provide OpenSubsonic compatibility for authorized libraries while the native Music API/internal architecture remain authoritative. | Medium | Planned |
| FR-040 | Implement a first-party `/api/v1/` Music API for modern auth, rich metadata, playback, offline sync, recommendations, radio, source/availability identity, and future features. | High | Planned |
| FR-041 | Deliver responsive web/PWA behavior where appropriate with keyboard/media-key support, notifications, playlist management, and adaptive Glaze UI. | High | Planned / mandatory delivery |
| FR-042 | Deliver a supported native Linux client/package with desktop integration, media keys, floating player, notifications, and release/rollback paths. | High | Planned / mandatory delivery |
| FR-043 | Deliver a first-class native Android app with background playback, media/lock-screen controls, offline downloads, sharing, widgets, dynamic Now Playing, and player/lyrics/queue/Credits gestures. | High | Planned / mandatory delivery |
| FR-044 | Plan a native iOS app under the same Music, privacy, security, source identity, and current-Stable Glaze UI contracts. | Medium | Planned |
| FR-045 | Implement Android Auto and plan CarPlay with driving-appropriate browsing and access to favorites, playlists, recent/downloaded music, radio, and recommendations. | Medium | Planned |
| FR-046 | Implement Android TV / Google TV and evaluate later Apple TV, Roku, Chromecast-compatible, AirPlay-compatible, and DLNA/UPnP ecosystems. | Medium | Planned / future expansion |
| FR-047 | Implement device/output management, playback transfer, device-specific settings/quality, and future multi-room playback where synchronization is supported. | Medium | Planned |
| FR-048 | Support individually controllable optional scrobbling, Discord Rich Presence, artwork, metadata, lyrics, discovery, and recommendation integrations without mandatory third-party dependencies. | Medium | Planned |
| FR-049 | Enforce privacy-by-default: no ad identifiers, no required third-party tracking, minimized telemetry, provider controls, private listening data, operational/history separation, and local recommendations by default. | High | Planned / mandatory platform contract |
| FR-050 | Enforce Wardveil Security for authentication, sessions, authorization, isolation, administration, sharing, downloads, abuse controls, secret handling, safe failure, and release evidence. | High | Planned / mandatory platform contract |
| FR-051 | Integrate GoreeCloud Identity, Search, Notify, Manager, Everkeep, Wardveil Security, Privacy Shield, and Glaze UI through authoritative contracts without bypassing Music authorization/privacy. | High | Planned / platform integration |
| FR-052 | Provide supported Docker and Podman/OCI delivery for self-hosted/server components with documented storage, backup, upgrade, rollback, monitoring, and recovery. | High | Planned / mandatory delivery |
| FR-053 | Keep source and availability as independent extensible track metadata so future providers do not require redesigning the user-facing state model. | High | Planned / governing data model |
| FR-054 | Keep proposed YouTube integration optional and gated by technical feasibility, provider policy, licensing/authorization, privacy, security, legal review, failure behavior, and replacement-path documentation. | High | Proposed; approval required |
| FR-055 | Maintain current-Stable Glaze UI conformance across every GoreeCloud-controlled Music interface as a mandatory Stable gate. | High | Ongoing release control |
| FR-056 | Maintain Everkeep-aligned backup/recovery/preservation for required application state while original music files remain independent from application database state. | High | Planned / mandatory platform contract |
| FR-057 | Define canonical Recording, Release, Source Item, and Playable Asset identities so routing cannot silently substitute remasters, live versions, edits, covers, or other materially different recordings. | High | Planned / governing data model |
| FR-058 | Implement match confidence such as Exact, Equivalent, Alternate, and Unknown; automatic rerouting is limited to Exact unless the user explicitly authorizes broader substitution. | High | Planned |
| FR-059 | Implement a generic provider-adapter contract for provider identity, search/browse, playback, seek/range behavior, quality, metadata/artwork, lyrics, authentication, quota/rate behavior, health, and offline capability. | High | Planned |
| FR-060 | Keep proposed YouTube support behind the same provider-adapter contract with no privileged bypass around Music authorization, privacy, source identity, or download rules. | High | Proposed; approval required |
| FR-061 | Implement a concrete availability state machine whose derived user-facing states remain Online, Offline, Downloaded, Cached, or Unavailable while source identity remains independent. | High | Planned |
| FR-062 | Differentiate durable Downloaded assets from temporary Cached assets and remove offline-ready claims when a local asset is corrupt, incomplete, expired, or unauthorized. | High | Planned |
| FR-063 | Implement deterministic playback routing using queue/source intent, exact identity, local availability, authorization, user source preference, network state, provider capability, and quality policy. | High | Planned |
| FR-064 | Explain playback-routing and quality decisions to users, such as downloaded copy selected, preferred server source, explicit provider source, source unavailable, or transcoding required. | Medium | Planned |
| FR-065 | Persist queue entries with stable queue identity, recording identity, source-item identity, requested source, selected route, and route reason while rechecking authorization at playback. | High | Planned |
| FR-066 | Implement source-preserving federated search with bounded provider timeouts, local-first/privacy-safe ranking, graceful partial results, and grouped-but-inspectable equivalent variants. | High | Planned |
| FR-067 | Implement offline-asset manifests with source/recording/asset identity, codec/quality, size, integrity verification, owner profile, manual-vs-smart attribution, and lifecycle state without reusable secrets. | High | Planned |
| FR-068 | Implement integrity/lifecycle management for downloaded and cached assets, including eviction, corruption handling, storage protection, and required revalidation. | High | Planned |
| FR-069 | Implement transparent quality negotiation across user profile, client capability, source capability, connection policy, server load, and transcode capability while avoiding unnecessary transcoding. | Medium | Planned |
| FR-070 | Implement multi-device synchronization/playback continuity for user metadata, playlists, preferences, queue state, and applicable position/history while downloads remain device-local by default. | High | Planned |
| FR-071 | Add privacy-preserving Private Listening / Private Session mode that can exclude eligible playback from recommendations and ordinary history without bypassing required security/operational records. | Medium | Planned |
| FR-072 | Design Manager observability around scanner health, storage, stream counts, aggregate transcode load, provider/cache/download health, API latency, and errors without exposing private listening content by default. | High | Planned / privacy requirement |
| FR-073 | Implement source-scoped degraded modes so provider outages/auth failures/rate limits/policy changes do not block local library browsing, authorized offline playback, or unrelated first-party capabilities. | High | Planned |
| FR-074 | Define stable native API domains for libraries, recordings, releases, source items, sources, search, playback sessions, queues, playlists, downloads, recommendations, radio, lyrics, devices, history, and sharing. | High | Planned |
| FR-075 | Adopt the documented Milestone 0–8 implementation sequence and require verified source/test evidence before any milestone, feature, or client target is treated as complete. | High | In progress — Milestone 0 is merged. Milestone 1 now has verified merged persistence/authorization and source-preserving scanner/reconciliation foundations; PR #9 is authoritative on `main` at `ac42ebc6f5fe3143c0cbc77cd6c3fce7397f44ac` with post-merge CI `35020139101` and Platform Contract `35020140030` passed. The remainder of Milestone 1 and Milestones 2–8 remain pending. |
| FR-076 | Maintain an acceptance matrix covering multi-user isolation, source disablement, provider failure, offline operation, queue recovery, exact-match routing, cache/download integrity, sync conflicts, accessibility, security/privacy, backup/recovery, and current-Stable Glaze UI. | High | Planned / verification control |

## Maintenance and synchronization

This roadmap and the Drive `FEATURE-ROADMAP.docx` must remain materially synchronized with each other and with the authoritative product specification. Update both copies whenever feature scope, priority, dependency, implementation status, cancellation, supersession, recommendation, or verification state materially changes.

No feature may be represented as complete or Stable solely because it appears here. Completion and lifecycle claims require applicable authoritative implementation, validation, review, release, production, security, privacy, recovery, and current-Stable Glaze UI evidence.

## Reconciliation rule

At each material feature change, reconcile this roadmap against the current authoritative project record, repository implementation state, applicable platform-system requirements, and GoreeCloud Tasks Management.
