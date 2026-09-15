# GoreeCloud Music — Repository Specifications

## Authority and status

This repository is the canonical source repository for **GoreeCloud Music**. The authoritative product scope is maintained in `Project Specification — Music.docx` in GoreeCloud Google Drive. This file is the repository-side engineering summary and must remain consistent with that record and with current cross-project GoreeCloud platform governance.

**Lifecycle:** Development  
**Current implementation status:** documentation and architecture foundation; no production-accepted Music runtime is established by this file  
**Capability identity:** GoreeCloud Resonance  
**Design system requirement:** latest approved Stable Glaze UI; current Stable is **1.4.1**  
**Platform Contract requirement:** current Contract **0.3**, evaluating exactly eight Integral Platform Systems  
**Required Integral Platform Systems:** GoreeCloud Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze UI, GoreeCloud Mesh, GoreeCloud Identity, and GoreeCloud Sync

## Product direction

GoreeCloud Music is an original GoreeCloud-owned, self-hosted-first, multi-user music platform. Navidrome and OpenSubsonic may be used as interoperability, migration, protocol, and behavioral references, but they do not define the permanent application architecture.

The product must preserve a two-dimensional media state model:

- **Source identity** answers where a playable item comes from, such as GoreeCloud Server, Internet Radio, or a separately approved external provider.
- **Availability** answers whether that item can play now without reaching its source, using states such as Online, Offline, Downloaded, Cached, and Unavailable.

Source identity must travel with queues, playlists, history, downloads, playback requests, and handoff state. Automatic routing must not silently substitute a materially different recording.

## Native API domains

The planned first-party `/api/v1/` surface should be organized around stable Music domains rather than provider-specific client contracts. Planned domains include libraries, recordings, releases, source items, sources, search, playback sessions, queues, playlists, downloads, recommendations, radio, lyrics, devices, history, and sharing.

## Multi-user and authorization model

Each user requires independent identity, permissions, library access, history, recommendations, queues, downloads, settings, devices, and private-data boundaries. Shared and family experiences must be explicit and must not silently expose private listening history or grant access to otherwise inaccessible media.

Runtime authorization must be rechecked at sensitive operation boundaries. Offline authorization may use a deliberately bounded local cache where approved, but offline operation must not become an unlimited bypass around revocation, profile separation, or device security.

## Synchronization boundary

GoreeCloud Sync is a required Platform System evaluation for Music. Planned cross-device queues, progress, playlists, downloads, preferences, and other synchronized state must use an explicit authorized dataset and conflict/reconciliation model rather than treating network reachability or shared storage as synchronization acceptance.

No Sync runtime integration or acceptance is established by the current repository baseline. The Platform Contract must remain fail-closed until implementation and evidence exist.

## Provider boundary

External provider support is optional. Provider adapters must use a generic contract for identity, discovery/search, playback capability, metadata, authorization, quota/rate behavior, health, quality, and offline capability.

The proposed YouTube integration is **not approved merely by appearing in product planning**. Any YouTube adapter or download behavior remains gated by provider policy, technical feasibility, licensing/authorization, privacy, security, legal review, and explicit product approval.

## Delivery targets

Planned delivery includes responsive web/PWA, Linux, Android, later iOS, automotive and living-room surfaces, and a self-hosted/server component where applicable. A planned target is not a statement of current support.

## Milestones

- **Milestone 0:** architecture, domain identity, API contracts, authorization model, provider-adapter contract, storage model, and automated test foundation.
- **Milestone 1:** native multi-user library service, scanner/indexing, metadata/artwork, library authorization, favorites/ratings, and persistent first-party state.
- **Milestone 2:** first-party playback, original/direct playback, transcoding negotiation, queues, source/availability metadata, and core `/api/v1/` behavior.
- **Milestone 3:** first-class Glaze UI web, Android, and Linux clients.
- **Milestone 4:** offline downloads, storage management, multi-device synchronization, continuity, and recovery behavior.
- **Milestone 5:** local recommendations, radio, lyrics, credits/metadata, private history/analytics, Recap, and authorized family/collaboration capabilities.
- **Milestone 6:** generic external-provider adapters and separately approved provider integrations.
- **Milestone 7:** automotive, television, output transfer, approved casting/receiver ecosystems, and future multi-room capabilities.
- **Milestone 8:** stabilization, accessibility, performance, security/privacy, recovery, migration, current-Stable Glaze UI conformance, and production-readiness review.

Milestone completion requires verified source and test evidence. Documentation, generated clients, prototypes, passing builds, or provider proofs of concept do not by themselves establish milestone completion or Stable status.
