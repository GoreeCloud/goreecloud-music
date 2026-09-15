# GoreeCloud Music — Features and Implementation State

## Implemented in the Milestone 0 foundation

- Typed domain identities for recordings, releases, source items, playable assets, queue items, and profiles.
- Source-kind and match-confidence models.
- Derived availability state with independent source identity.
- Provider adapter and registry contracts.
- Authorization request/decision contract.
- Deterministic playback route-selection core with no automatic non-exact substitution by default.
- Stable queue-item model carrying recording/source/route context.
- Minimal development service with health and build-information endpoints.
- OpenAPI development contract for implemented endpoints plus separately declared planned API domains.
- Machine-readable GoreeCloud Platform Contract state that does not claim unimplemented integrations.
- Unit tests and GitHub Actions CI definition.

## Partial / foundation only

- Playback routing: decision core exists; actual stream acquisition, codec negotiation, transcoding, and sessions do not.
- Provider architecture: contract and registry exist; no real provider adapters exist.
- Authorization: contract exists; GoreeCloud Identity/Privacy Shield/Wardveil runtime integration does not.
- API: service shell and two operational endpoints exist; product APIs are not implemented.
- Queue: domain model exists; persistence, multi-device recovery, and shared queues do not.

## Planned

All other items remain governed by `FEATURE-ROADMAP.md` and the authoritative project specification. Roadmap presence is not implementation evidence.
