# GoreeCloud Music — Repository Specifications

**Repository version:** `0.1.0-dev.1`  
**Lifecycle:** Active Development  
**Current milestone:** Milestone 0 — Architecture Foundation

This file records repository-coupled implementation specifications. The authoritative product scope and planned capability direction remain in `GoreeCloud/Projects/Project Specification — Music.docx`.

## Native architecture

GoreeCloud Music is original GoreeCloud-owned software. Whole-product Navidrome or other third-party application architecture is not the implementation foundation. OpenSubsonic/Navidrome compatibility may be added later through bounded interoperability layers.

## Milestone 0 contracts

The current implementation establishes:

- stable domain identity types for Recording, Release, Source Item, Playable Asset, Queue Item, and Profile;
- source identity independent from availability state;
- Exact, Equivalent, Alternate, and Unknown match confidence;
- provider-neutral adapter interfaces;
- application authorization decision interface;
- deterministic playback route selection with exact-match protection;
- stable queue identity fields;
- minimal native service identity/health endpoints;
- an OpenAPI development contract that separates implemented paths from planned domains;
- truthful machine-readable Platform-System status;
- baseline tests and CI.

## Status boundary

These contracts do not establish production authentication, actual music playback, persistent library storage, scanning, external provider support, user-facing clients, or Stable qualification. Those remain later milestone obligations.
