# GoreeCloud Music — Planned Features and Capabilities

> **Repository mirror only.** The authoritative product record is `GoreeCloud/Projects/Project Specification — Music.docx`. This Markdown file exists for source-controlled continuity and must remain materially synchronized with the authoritative Drive specification. Planned or proposed capability text is not implementation evidence.

**Version:** v0.2  
**Status:** Active Development — planned capability catalog and implementation architecture established; provider-specific expansion remains proposed unless separately approved and verified  
**Application:** GoreeCloud Music  
**Capability identity:** GoreeCloud Resonance  
**Canonical repository:** `GoreeCloud/goreecloud-music`  
**Design system:** latest approved Stable Glaze UI release  
**Security:** Wardveil Security  
**Privacy:** Privacy Shield  
**Continuity:** Everkeep

## Current reconciliation — September 15, 2026

This specification mirror records the same product direction as the authoritative Drive specification. The canonical repository was newly initialized on September 15, 2026 and does not by itself provide implementation evidence sufficient to treat the features below as implemented. The YouTube source integration remains a proposed optional expansion subject to technical, provider-policy, licensing, privacy, security, authorization, and legal review.

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

- Prefer an exact authorized downloaded copy, then authorized server copy, then approved online source according to user preference.
- Preferences include Prefer Local / Offline, Prefer GoreeCloud Server, and Ask When Multiple Sources Exist.
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

# Implementation Architecture Continuation

The following architecture extends the planned capability catalog into an implementation model. It remains planned architecture and does not claim implementation.

## Core domain and authority boundaries

GoreeCloud Music Server should remain authoritative for first-party libraries, user music permissions, Music profiles, playlists, favorites, ratings, playback history, recommendation state, queues, source mappings, downloads, and application-specific settings. External providers are subordinate integrations rather than alternate application authorities.

- Clients should consume the native Music API rather than embed provider-specific business logic throughout the UI.
- Provider adapters should return capability/evidence to the Music domain, which applies authorization, routing, privacy, and state rules.
- OpenSubsonic compatibility should remain an adapter boundary rather than dictate the internal model.
- Optional providers must be removable without making the self-hosted library unusable.

## Canonical recording and asset identity

The source model should distinguish four durable concepts:

- **Recording Identity** — the specific performance/recording the user intends to hear.
- **Release Identity** — the album, single, compilation, edition, remaster, or other release containing that recording.
- **Source Item Identity** — the provider-specific object representing that recording/release.
- **Playable Asset Identity** — a concrete encoded asset or stream, including codec, quality, version, and availability characteristics.

Matching should have an explicit confidence classification such as **Exact**, **Equivalent**, **Alternate**, and **Unknown**. Automatic source rerouting should normally be limited to Exact matches. A remaster, live performance, radio edit, clean/explicit variant, cover, remix, or materially different duration must not be silently substituted merely because artist/title text resembles the requested track.

## Source-provider adapter contract

A generic source adapter should expose provider identity and capabilities through one consistent contract. Candidate domains include:

- provider identity/version and health;
- search and browse;
- metadata and artwork;
- playback URL/session acquisition;
- seek/range and resume behavior;
- available codec/quality profiles;
- authentication/authorization requirements;
- provider quota/rate-limit information;
- lyrics or timed-text capability where applicable;
- offline/download capability declaration;
- failure classification and retry guidance.

The proposed YouTube integration must use this same adapter contract. It must not receive a privileged bypass around Music authorization, privacy rules, source identity, playback routing, or download policy.

## Availability state machine and cache semantics

The user-facing states remain Online, Offline, Downloaded, Cached, and Unavailable, but they should be derived from concrete runtime state rather than stored as vague labels.

- **Downloaded** means a deliberately retained local asset is complete, integrity-verified, authorized for the current user/device, and intended for offline playback.
- **Cached** means a temporary local playback optimization that may be evicted and must not be presented as a durable offline promise.
- **Offline** means playback can succeed without contacting the originating source at that moment.
- **Online** means source access is presently required and believed available.
- **Unavailable** means the requested source/asset cannot currently be played under the applicable authorization/capability state.

Corrupt, incomplete, expired, or unauthorized local material must lose any offline-ready claim. Source identity remains unchanged when availability changes.

## Deterministic playback decision engine

Playback routing should be deterministic, inspectable, and testable. The decision engine should consider, in order appropriate to policy:

1. the queue item and requested recording/source intent;
2. exact recording/source identity;
3. authorized verified local downloads;
4. user source preference;
5. first-party server availability/authorization;
6. approved external-provider capability/authorization;
7. network state and connection policy;
8. client codec/device capability;
9. quality profile and transcode policy;
10. explicit user confirmation where substitution is not Exact.

The selected route should carry an explanation such as **Downloaded copy selected**, **Preferred GoreeCloud Server source**, **Explicit provider source**, **Source unavailable**, or **Transcoding required**. This enables UI explanation, debugging, and acceptance tests without exposing sensitive internals.

## Queue and playback-session integrity

A queue entry should persist stable identity rather than only a display title and URL. Candidate fields include queue item ID, recording identity, source item identity, requested source, selected route, route reason, position/order, user/profile owner, and revision information.

- Authorization must be checked again when playback starts; queue persistence is not perpetual authorization.
- Provider failure should not rewrite the queue entry into another recording.
- Queue/session recovery should preserve identity after application restart, device handoff, temporary provider outage, or transient network loss.
- Shared/listen-together queues must preserve participant/host permissions and must not expose private source credentials.

## Federated search and source-preserving deduplication

Search should be federated without pretending that equivalent provider results are the same stored object.

- Local/private results should be independently queryable and may receive preference when configured.
- Online providers should use bounded timeouts and graceful partial-result behavior.
- Equivalent results may be visually grouped, but each source variant must remain inspectable and selectable.
- Ranking should avoid sending unnecessary private-library/history data to external providers.
- Provider search failures must not make first-party search fail.

## Offline asset manifests and integrity

Each retained offline asset should have a durable manifest recording at least source identity, recording identity, playable asset identity, codec/quality, expected size, integrity/checksum data where practical, owning profile, manual-versus-Smart-Download attribution, lifecycle state, and any provider-defined expiry/revalidation condition. Reusable provider credentials or secrets must not be embedded in the manifest.

Lifecycle handling should define complete, validating, available, expired, corrupt, evicted, and removal states. Storage quotas and Smart Downloads should never evict protected/manual items unless the user explicitly authorizes that behavior.

## Playback quality and transcoding negotiation

Quality selection should resolve transparently across user quality profile, device/client codec support, source asset capability, connection policy, server transcode capability/load, and accessibility or platform constraints.

- Prefer direct/original playback where allowed and practical.
- Avoid transcoding merely because it is available.
- Expose original-vs-transcoded state, source codec/quality, delivered codec/quality, and the reason for material quality changes where known.
- Quality decisions should remain independent from source identity.

## Multi-device synchronization and playback continuity

Sync should use stable domain identifiers and revision information rather than overwrite-by-timestamp alone. Synchronizable state may include favorites, ratings, playlists, applicable history/position, playback preferences, device capability/preferences, and queue state.

Downloads remain device-local by default. A server may synchronize the intent/state that an item is saved, but another device should not falsely display that asset as locally Downloaded until it actually possesses and validates its own offline copy.

Conflict rules should preserve user edits where practical. Playback transfer must not expose another profile’s private queue/history/source state.

## Recommendation, history, and analytics privacy boundary

Recommendation state and ordinary listening analytics remain per-user by default.

- Recommendation events should be separable from operational telemetry.
- Users should be able to clear/reset eligible recommendation/history state subject to retention/recovery rules.
- A planned **Private Listening / Private Session** mode should allow eligible playback to avoid recommendations and ordinary history surfaces without bypassing required security/audit records.
- External recommendation providers, if approved, should receive minimized inputs and remain separately controllable.

## Administrative observability without listening surveillance

GoreeCloud Manager should focus on system health rather than exposing private listening content. Useful operational signals include scanner state/duration, storage usage, stream counts, aggregate transcode load, provider health, cache/download health, API latency, and error classes.

Logs should prefer opaque identifiers and technical context over raw private metadata. Content-level diagnostics should require a legitimate troubleshooting purpose, narrow scope, authorization, and appropriate privacy handling.

## Failure isolation and degraded-mode behavior

Optional-provider failures should be source-scoped.

- Provider outage must not prevent local library browsing or authorized offline playback.
- Provider authentication failure should limit that provider until reauthorization.
- Rate-limit or policy failures should not disable unrelated first-party capabilities.
- Queue entries retain identity so they can recover later or be explicitly rerouted.
- Offline-capable clients may use an approved local authorization cache for authorized offline playback, but that cache must not become a general bypass around revocation/expiry rules.
- When connectivity returns, synchronization and provider/session authorization should be revalidated.

## Native API domains

Candidate stable `/api/v1/` application domains include:

- libraries;
- recordings;
- releases;
- source items and sources;
- search;
- playback sessions;
- queues;
- playlists;
- downloads/offline assets;
- recommendations and radio;
- lyrics and credits;
- devices/output;
- history/analytics;
- sharing/collaboration.

Provider-specific details should remain behind adapter/domain boundaries rather than leak into every first-party client contract.

## Milestone sequence

**Milestone 0 — Architecture foundation.** Domain identities, API contracts, authorization model, provider-adapter contract, storage model, migrations, baseline tests, CI, and current platform-system integration plan.

**Milestone 1 — Native multi-user library.** Library service, scanning/indexing, metadata/artwork, library authorization, favorites/ratings, recently added/played, and isolation tests.

**Milestone 2 — First-party playback.** Direct/original playback, transcode negotiation, durable queues, source/availability metadata, playback sessions, and core `/api/v1/` contracts.

**Milestone 3 — First-class clients.** Glaze UI web, Android, and Linux clients with Now Playing, search, library browsing, playlists, queue, source/availability state, and accessibility acceptance.

**Milestone 4 — Offline and continuity.** Downloads, storage manager, Smart Downloads foundation, multi-device synchronization, playback continuity, offline authorization/recovery behavior, and integrity tests.

**Milestone 5 — Resonance intelligence and collaboration.** Local recommendations, radio, lyrics, richer Credits/metadata, history/analytics, Recap, Listen Together, and family/collaboration controls.

**Milestone 6 — External provider framework.** Generic provider adapter plus separately approved provider integrations. Proposed YouTube functionality cannot bypass provider-policy, legal, authorization, privacy, security, or offline-copy approval gates.

**Milestone 7 — Extended playback ecosystems.** Automotive, television, output transfer, approved casting/receiver ecosystems, and future multi-room capability where synchronization can be proved.

**Milestone 8 — Stabilization and release acceptance.** Accessibility, performance, security/privacy validation, backup/recovery validation, migration, client packaging/signing, current-Stable Glaze UI conformance, and production-readiness evidence.

Milestone completion requires verified source/test evidence. Passing builds, prototypes, generated artifacts, roadmap entries, or documentation updates do not alone qualify a milestone as complete.

## Acceptance matrix requirement

Before Stable eligibility, acceptance evidence should cover at least multi-user isolation, source disablement, provider failure, offline operation, queue/session recovery, exact-match routing, cache/download integrity, synchronization conflict handling, accessibility, security/privacy, backup/recovery, and current-Stable Glaze UI conformance.

## Verification boundary

This file is documentation, not proof of implementation. Features and milestones must not be represented as implemented, complete, production-ready, or Stable without applicable authoritative source, test, review, release, runtime, security, privacy, recovery, and current-Stable Glaze UI evidence.
