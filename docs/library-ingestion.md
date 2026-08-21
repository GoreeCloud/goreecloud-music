# Resonance Library ingestion

Resonance Library ingestion turns ordinary music files on configured storage into GoreeCloud Music database records without modifying the source collection.

## Scan flow

An authorized library manager can request:

```text
POST /api/v1/libraries/{libraryID}/scan
```

The server resolves the library, verifies the persisted request principal, enforces library-management permission, recursively discovers supported audio files beneath the configured root, probes each candidate locally with `ffprobe`, and reconciles successful metadata into PostgreSQL.

Administrators may manage library scans through the existing administrative authorization policy. Non-administrators require a matching library membership with `can_manage=true`.

## Source-file safety

The scanner is read-only with respect to the music collection. It does not rename, move, rewrite, retag, truncate, or delete source audio files.

A complete successful scan records a single scan-start timestamp. Files seen during that scan receive that timestamp. Only after every discovered candidate has been probed and persisted successfully may database records for previously indexed files not seen during the current scan be removed. Orphaned database track records are then cleaned up.

If any probe or persistence operation fails, the scan is reported as incomplete and stale-record cleanup is skipped. This prevents a partial scan, temporary storage problem, unsupported/corrupt file, or metadata failure from being interpreted as authoritative evidence that previously indexed music disappeared.

## Metadata

The ingestion foundation currently records or derives:

- title, with filename-stem fallback
- artist
- album
- album artist
- genre
- release year
- track number
- disc number
- duration
- codec
- bitrate
- bit depth
- sample rate
- channels
- file size

Track, disc, and year values accept common leading-number forms such as `4/10`, `1/2`, and `2026-08-21`.

## Reconciliation boundaries

Track-file path is the current stable reconciliation key. A path already associated with another library is rejected rather than silently reassigned. Artist records are reused case-insensitively, and album records are reused within a library by title and album-artist relationship.

This is a Milestone 1 ingestion foundation. Artwork ingestion, stronger media fingerprints, rename/move detection, batch/background scan scheduling, richer multi-artist modeling, and production operational controls remain future work.

## Runtime dependencies

The current local metadata provider requires `ffprobe` to be installed wherever the GoreeCloud Music API performs scanning. External metadata services are not required for this ingestion path.
