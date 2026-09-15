# Media Integrity Hashing

## Status

Development-only Milestone 1 foundation.

This boundary records SHA-256 over the exact bytes of an already-materialized GoreeCloud Server playable asset. It uses the existing nullable `playable_assets.media_sha256` field and does not introduce a schema version or third-party dependency.

## Preconditions

`HashLibraryFileContent` requires:

- a validated requesting Profile ID and Library ID;
- current library edit or owner permission;
- a durable scanner `file_id` that is not tombstoned missing;
- the deterministic GoreeCloud Server Source Item created for that library file;
- the deterministic Playable Asset created for that library file;
- exact agreement among scanner relative path/size, Source Item recording binding, Playable Asset recording/source/path/size binding, and the rooted source file;
- a source file that still matches the scanner size/mtime observation before and after hashing.

The source file is opened through the same rooted, symbolic-link-rejecting, regular-file boundary used by metadata extraction and format probing. Hashing is read-only and checks context cancellation while streaming the file.

## Persisted fact

The operation streams the complete source file through SHA-256 and persists the lowercase hexadecimal digest in `playable_assets.media_sha256` only after source binding and freshness checks succeed. The update is performed inside the same serializable transaction used to re-evaluate authorization and application-state bindings.

Exact retries over unchanged source bytes are idempotent. If the source has changed relative to the scanner observation, the operation fails and the previous stored digest remains untouched.

## Identity boundary

`media_sha256` identifies exact file bytes. It is **not** canonical Recording or Release identity and must not be used by itself to conclude that two files are the same musical recording, edition, master, performance, or user-intended source.

Two files with different tags or container bytes can represent the same recording while producing different hashes. Conversely, a matching file hash proves byte equality for the hashed files, not broader rights, authorization, provider equivalence, or playback policy.

## Security and integrity boundary

The digest can support later corruption detection and integrity revalidation, but this slice does not yet define a complete integrity lifecycle, trusted baseline, offline-asset manifest, recovery policy, or periodic verification scheduler. Scanner size/mtime checks remain freshness guards around the hashing operation; they are not cryptographic identity.

## Explicitly deferred

This foundation does not implement:

- automatic Recording/Release matching or duplicate-equivalence decisions;
- audio-essence fingerprints or acoustic matching;
- periodic/background integrity verification;
- offline/download manifest integrity policy;
- recovery trust anchors or Everkeep restore acceptance;
- duration, bitrate, sample rate, channels, bit depth, or ReplayGain probing;
- broader format probing;
- public API or UI exposure;
- production qualification, release eligibility, or Stable status.
