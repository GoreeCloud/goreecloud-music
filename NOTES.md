# GoreeCloud Music — Notes

## Current repository state

- Lifecycle: **Development**.
- Canonical product record: `Project Specification — Music.docx` in GoreeCloud Google Drive.
- Canonical feature control: repository `FEATURE-ROADMAP.md` paired with Drive `GoreeCloud/Feature Roadmap/GoreeCloud Music/FEATURE-ROADMAP.docx`.
- Capability identity: **GoreeCloud Resonance**.
- Current Platform Contract requirement: **0.3**, with exactly eight Integral Platform Systems evaluated.
- Current Stable design-system requirement: **GLAZE UI V1.4.1 / 1.4.1**.
- No supported production Music runtime, release, deployment, or Stable acceptance is established by the current repository evidence.

## Architecture direction

GoreeCloud Music is original GoreeCloud-owned software. Navidrome and OpenSubsonic are references and interoperability/migration targets rather than permanent complete-application foundations.

Source identity and availability are separate concepts. Recording identity must remain stable enough to prevent silent substitution between materially different recordings.

External providers use a generic provider-adapter boundary. The proposed YouTube integration remains separately gated and is not authorized merely because it appears in planning or roadmap documents.

GoreeCloud Sync must be treated as its own authorization and state-reconciliation boundary for any future cross-device Music state. Mesh reachability, shared storage, downloads, or Everkeep recovery are not substitutes for accepted Sync behavior.

## Current improvement branch

`agent/music-repository-baseline-20260915` adds the missing repository-governance controls required for the GoreeCloud application/service baseline, migrates the repository declaration to Platform Contract 0.3, explicitly evaluates GoreeCloud Sync, and adds exact-revision CI. It deliberately does not manufacture executable product implementation, runtime platform-system acceptance, or lifecycle promotion.

## Immediate next engineering priorities

1. Establish Milestone 0 executable architecture, domain packages, API contracts, storage model, authorization model, provider-adapter interface, and automated test foundation.
2. Establish repository CI appropriate to the selected implementation languages and clients, while retaining the current governance/Platform Contract validation gate.
3. Implement the native multi-user library foundation before provider-specific expansion.
4. Add evidence-backed Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze UI, GoreeCloud Mesh, GoreeCloud Identity, and GoreeCloud Sync integration as implementation becomes real.
5. Define authorized Sync datasets, version/conflict behavior, offline reconciliation, and revocation behavior before representing cross-device state as synchronized.
6. Add supported build, packaging, accessibility, rollback, recovery, and release evidence before any lifecycle promotion.
7. Keep repository roadmap, Drive roadmap, project specification, Suite inventory, and Tasks Management synchronized with verified state.

## Evidence boundary

Documentation describes intended architecture and product obligations. Completion requires verified source, tests, builds, integration acceptance, release evidence, and runtime evidence appropriate to the claim.
