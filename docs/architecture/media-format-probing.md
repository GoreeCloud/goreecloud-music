# GoreeCloud Music — Bounded Media Format Probing

Status: Development foundation

## Purpose

This Milestone 1 slice adds truthful codec/container facts for already-materialized GoreeCloud Server playable assets without expanding canonical identity or playback claims.

## Boundary

`ProbeLibraryFileFormat` operates only after explicit-identity ingestion has produced the deterministic Source Item and Playable Asset for a durable scanner `file_id`.

The operation:

- re-evaluates current library edit/owner authorization;
- requires the current scanner file to be present and still bound to the expected source item, playable asset, media path, and size;
- opens the source through the existing rooted, non-symlink file boundary;
- validates actual source bytes rather than trusting the filename extension;
- verifies the scanner size/mtime snapshot again after probing;
- persists only `codec` and `container` on the existing playable asset;
- is idempotent for unchanged source state;
- leaves existing format state untouched when probing fails.

## Current bounded support

### MP3

The `.mp3` parser accepts ID3v2.3/v2.4-prefixed or raw MPEG audio only when it can confirm two consecutive MPEG Layer III frame headers with compatible MPEG version and sample rate. A tag-only `.mp3` file does not qualify as MP3 media.

The persisted facts are:

- codec: `mp3`
- container: `mpeg-audio`

### FLAC

The `.flac` parser requires the native `fLaC` signature and a first STREAMINFO metadata block with the required 34-byte length.

The persisted facts are:

- codec: `flac`
- container: `flac`

## Explicitly not established

This slice does not establish or persist:

- duration;
- bitrate;
- sample rate;
- channel count;
- bit depth;
- ReplayGain;
- media hashes or integrity qualification;
- artwork;
- sidecar metadata;
- additional audio formats;
- automatic Recording/Release identity matching or duplicate equivalence;
- background ingestion/probing;
- public API or client behavior;
- playback/transcoding acceptance;
- production recovery, release eligibility, Platform-System runtime acceptance, or Stable qualification.

Those remain separate Milestone 1/later release obligations and require their own verified evidence.
