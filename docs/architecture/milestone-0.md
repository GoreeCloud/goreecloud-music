# Milestone 0 — Architecture Foundation

## Status

Implemented foundation on the Milestone 0 development line. This is not a Stable or production-readiness claim.

## Implemented boundaries

- Typed Recording, Release, Source Item, Playable Asset, Queue Item, and Profile identities.
- Independent source identity and derived availability state.
- Exact/equivalent/alternate/unknown match confidence model.
- Provider-neutral adapter interface and registry.
- Application authorization contract.
- Deterministic route selection that refuses silent non-exact substitution by default.
- Stable queue identity model with owner/source/recording/route fields.
- Minimal native HTTP service exposing health and development build identity.
- API contract that labels implemented paths and lists future API domains separately.
- Machine-readable Platform-System state that truthfully reports integrations as planned rather than complete.
- Baseline Go tests and CI workflow.

## Explicitly not implemented in Milestone 0

Library persistence/scanning, user authentication, real provider integrations, media streaming/transcoding, playlists, downloads, recommendations, synchronization, clients, Glaze UI surfaces, Platform-System runtime integrations, deployment packaging, and production recovery are not implemented by this milestone.

## Authority boundaries

External source providers remain subordinate adapters. They do not become authorization, privacy, security, recovery, or application-state authorities. Queue/source identities survive provider failure; route selection may change only within the explicit match and authorization policy.

## Persistence and migrations

A production persistence engine and migration format are intentionally not selected in this milestone. That decision remains an open implementation task so the repository does not create a speculative database dependency before library/account/storage requirements are validated.
