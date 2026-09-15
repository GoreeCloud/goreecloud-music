# GoreeCloud Music — Bounded Media Content Hashing

Status: Development foundation

## Purpose

This Milestone 1 slice adds a first-party SHA-256 content-hash path for an already-materialized GoreeCloud Server playable asset. The hash is durable source evidence. It is not production integrity qualification, a digital-signature scheme, or proof that storage is immutable.

## Boundary

`HashLibraryFileContent` operates only after explicit-identity ingestion has produced the deterministic Source Item and Playable Asset for a durable scanner `file_id`.

The operation:

- re-evaluates current library edit/owner authorization inside a serializable transaction;
- requires current non-missing scanner state and exact source-item, recording, asset, path, and size binding;
- opens the source through the existing rooted, read-only, non-symlink file boundary;
- hashes the complete opened file twice with SHA-256;
- requires each pass to read exactly the scanner-observed size;
- revalidates size/mtime after each pass and immediately before commit;
- requires both independently computed digests to agree before persistence;
- persists the lowercase hexadecimal SHA-256 only when the existing asset hash is null;
- treats an identical stored hash as idempotent verification;
- fails closed rather than overwriting a conflicting stored hash.

Original media is never modified.

## Explicitly not established

This slice does not establish:

- production storage integrity or anti-tamper guarantees;
- cryptographic signatures, trusted timestamps, or remote attestation;
- background or incremental hashing policy;
- automatic invalidation/re-hashing after every filesystem event;
- chunk hashes, Merkle trees, deduplication, or content-addressed storage;
- automatic Recording/Release identity matching or duplicate equivalence;
- duration, bitrate, sample rate, channel count, bit depth, ReplayGain, or broader media probing;
- public API/client behavior;
- backup/recovery acceptance;
- release eligibility, Platform-System runtime acceptance, or Stable qualification.

Those remain separate Milestone 1 or later release obligations and require their own verified evidence.
