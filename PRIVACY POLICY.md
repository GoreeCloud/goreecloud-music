# GoreeCloud Music — Privacy Policy

## Status and authority

This repository policy defines GoreeCloud Music privacy requirements during **Development**. It is subordinate to the current authoritative GoreeCloud privacy and data-protection policies and Privacy Shield requirements. If a conflict exists, the higher-authority GoreeCloud requirement controls.

This file describes required product behavior; it is **not** evidence that a production runtime currently satisfies these requirements.

## Privacy principles

GoreeCloud Music must be private by default, purpose-limited, data-minimized, least-privilege, understandable, portable, and recoverable. Self-hosting does not by itself establish privacy.

Music should collect, process, expose, retain, synchronize, or disclose only the information necessary for an approved user or operational purpose. Information obtained for playback, library management, recommendations, diagnostics, sharing, or another purpose must not be silently reused for unrelated profiling or tracking.

## Music data categories

Depending on implemented and enabled capabilities, Music may process:

- User-owned audio files and authorized library paths.
- Track, album, artist, release, artwork, lyrics, credits, identifiers, and technical audio metadata.
- Account, profile, library-authorization, session, and device information.
- Favorites, ratings, playlists, queues, listening position, listening history, play/skip counts, recommendations, and user preferences.
- Download/cache manifests, storage usage, integrity state, and offline authorization state.
- Sharing, collaboration, Listen Together, and family authorization metadata.
- Provider configuration, scoped authorization state, and provider tokens when a separately approved external provider requires them.
- Operational logs, health state, aggregate performance metrics, error information, and security/audit events.

Operational information can still be personal or sensitive. It must not be treated as harmless merely because it is machine-generated.

## User and family isolation

Each user must have an independent identity and privacy boundary for libraries, history, recommendations, queues, downloads, preferences, devices, and other private state. Shared or family experiences must be explicitly authorized.

Family or infrastructure administration does not create unrestricted permission to inspect another user’s private listening history, playlists, searches, or media. Administrative access must be used only for an approved operational, custodial, security, recovery, legal, or other legitimate purpose.

## Search privacy

Private-library searches must remain within authorized GoreeCloud processing by default. A private-library query must not be sent to an external provider merely because a combined-search feature exists.

External search must require an explicit external scope or another approved, clearly disclosed combined-search behavior. The interface must identify which sources are queried.

## Recommendations, analytics, and Recap

Recommendations should be local or GoreeCloud-controlled by default and use only signals authorized for that user and purpose. One user’s activity must not silently influence another user’s profile.

Listening history, analytics, monthly/annual Recap, and similar personalization are private per-user data by default. Household or family aggregation requires an explicit permission model for the relevant participants.

A future Private Listening or Private Session capability may exclude eligible playback from ordinary recommendation/history surfaces, but it must not bypass required security, abuse-prevention, or narrowly scoped operational records.

## Telemetry and monitoring

Music must not require advertising identifiers, behavioral advertising, or third-party tracking for core operation.

Operational telemetry should minimize private listening content. GoreeCloud Manager or monitoring integrations should prefer service health, scanner state, storage status, aggregate stream/transcode load, provider health, cache/download integrity, latency, and error rates rather than exposing track titles, listening history, search terms, or user content by default.

## External providers

External providers are optional and must be separately enabled and authorized. Source identity must remain visible to the user. Provider access must be limited to the scopes and information required for the selected capability.

The proposed YouTube integration is not approved merely because it appears in planning. Provider policy, licensing/authorization, privacy, security, legal review, and technical feasibility must be satisfied before implementation or release.

Playback availability does not automatically authorize downloading, caching, copying, redistributing, or bypassing provider protections.

## Offline data

Downloaded and cached assets must preserve the owning profile, original source identity, authorization context, integrity state, and applicable lifecycle. Local assets should be protected at rest where appropriate for the target platform and sensitivity.

Offline behavior must not become an unlimited authorization bypass. Revocation, expiration, corruption, profile separation, and device-security conditions must be handled explicitly. When an asset is no longer valid for offline playback, the interface must stop representing it as available.

## Retention, deletion, export, and portability

Music should retain data only for a defined purpose and lifecycle. Users must have understandable controls or documented workflows for deleting eligible history, downloads, caches, playlists, preferences, and other user-controlled state.

Supported application state should be exportable in documented, portable forms where practical. Original user-owned media should remain independent from the Music application database so loss or replacement of application state does not imply loss of the source media library.

## Backup and recovery privacy

Backups and restores must preserve ownership, authorization, encryption, and privacy boundaries. Recovery must not expose one user’s data to another user or restore stale authorization as if it were current.

Everkeep-aligned recovery must be separately validated before production-recovery claims are made.

## Privacy Shield

Privacy-sensitive operations must integrate the applicable Privacy Shield contract. Authorization should travel with the operation and purpose rather than being inferred solely from a logged-in identity. Missing or unaccepted privacy authorization must fail closed where the operation requires it.

No Privacy Shield runtime acceptance is established by this repository baseline.