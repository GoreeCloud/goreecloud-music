# GoreeCloud Music — Security

## Current status

GoreeCloud Music is in **Development**. This repository baseline does not establish a production security review, Wardveil runtime acceptance, hardened deployment, supported release, or Stable status.

## Security principles

Music must use least privilege, explicit authorization, secure defaults, defense in depth, auditable state transitions, safe failure, and minimum necessary exposure. Authentication identifies a principal; it does not grant unrestricted media, administrative, provider, download, sharing, or family access.

Wardveil Security and GoreeCloud Identity must be integrated through their accepted contracts where applicable before production security claims are made.

## Primary trust boundaries

Security design and testing must explicitly cover:

- Account, profile, session, device, and library authorization.
- Library roots, filesystem scanning, imports, uploads, and path handling.
- Untrusted tags, metadata, artwork, lyrics, playlists, and sidecar files.
- Audio parsing, range requests, decoding, transcoding, and media workers.
- API authentication, authorization, pagination, filtering, rate limiting, and object-level access.
- Queues, playlists, shares, family access, collaboration, and Listen Together.
- Download/cache creation, integrity, ownership, expiration, revocation, and deletion.
- Provider adapters, external credentials, callback flows, quotas, and rate limits.
- Administrative and GoreeCloud Manager interfaces.
- Backup, restore, migration, export, and recovery paths.

## Filesystem and media ingestion

Library scanners and import paths must prevent traversal outside approved roots, unsafe symlink behavior, confused ownership, and unauthorized file discovery. A path being readable by the service account is not sufficient user authorization.

Untrusted metadata, artwork, lyrics, playlist content, and provider responses must be parsed defensively and safely encoded before being displayed or persisted. File type and content validation should not rely only on user-controlled extensions or metadata.

Transcoding and media-processing components should run with narrowly scoped filesystem, network, process, and resource permissions. Resource controls must prevent malformed or adversarial media from creating unbounded CPU, memory, disk, or process consumption.

## Authorization

Every sensitive read or mutation must be authorized against the current user/profile, resource, library, source, device, and operation where applicable. Authorization must be re-evaluated at playback time for queued items rather than assuming that earlier queue creation permanently grants access.

Shared playlists or collaboration metadata must not grant underlying media access implicitly. Listen Together participants must each satisfy the required access rules for private content.

Offline authorization may use an explicitly designed bounded cache when approved, but it must not defeat revocation, profile separation, device security, or expiration rules.

## Recording and routing integrity

Playback routing must not silently substitute a materially different recording merely because metadata strings are similar. Stable recording/source identity and explicit match confidence reduce spoofing, content confusion, and user-intent violations.

## Secrets and provider credentials

Passwords, API tokens, provider refresh tokens, signing secrets, encryption keys, recovery information, and other reusable credentials must never be committed to the repository or emitted in ordinary logs.

External provider credentials must use minimum required scopes and must be independently revocable. Provider failure or loss of authorization must degrade only the affected source where practical rather than disabling unrelated local capabilities.

## Logging and audit

Security and operational events should be sufficient to investigate authentication, authorization, sharing, provider, download, integrity, and administrative failures without unnecessarily recording private listening content. Logs must avoid raw credentials, reusable tokens, private media bytes, or unnecessary search/history details.

## Dependencies and supply chain

Dependencies must be minimized, pinned or constrained appropriately, reviewed for security and licensing risk, and updated through controlled source review. Mature security, cryptographic, codec, protocol, operating-system, and runtime components may remain narrow dependencies when independent reimplementation would materially increase risk.

Release preparation should include dependency and vulnerability review, secret scanning, build provenance, artifact integrity, and software-bill-of-materials controls where applicable.

## Required acceptance work

Before production or Stable qualification, Music requires evidence for representative authentication and authorization tests, multi-user isolation, path and parser safety, provider failure, rate limiting, download/cache integrity, security-sensitive logging, backup/restore authorization, supported-platform hardening, rollback/recovery, and current applicable Wardveil Security acceptance.

No such acceptance is implied by this document alone.