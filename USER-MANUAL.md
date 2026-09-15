# GoreeCloud Music — User Manual

## Current availability

GoreeCloud Music is in **Development**. This repository does not yet establish a production-accepted server or supported end-user client. The workflows below describe the intended product behavior and must not be read as instructions for a currently released service.

## Planned user model

Each user will have an independent GoreeCloud identity, authorized libraries, favorites, ratings, listening history, recommendations, queues, downloads, devices and settings. Shared or family access must be explicitly granted.

## Planned library workflow

1. An authorized user connects or selects a GoreeCloud Music library.
2. Music scans approved locations incrementally and reads supported embedded or sidecar metadata.
3. Library content appears under the user’s authorized scope.
4. Search, playlists, favorites, ratings and playback operate only against sources the user is allowed to access.

Original audio files are intended to remain independent from the application database so losing application state does not imply losing the media library.

## Source and availability

Music distinguishes **source** from **availability**.

A track may be sourced from GoreeCloud Server while being Online, Offline or Downloaded. Internet Radio is a separate source type. Optional external providers will be shown explicitly when approved and enabled.

If the requested recording has multiple sources, Music should follow the user’s source preference and recording identity. It should not silently replace a requested recording with a remaster, live version, cover or other materially different item.

## Planned playback controls

Now Playing is intended to provide artwork, source, availability, quality, transport controls, queue access, lyrics, Credits, output-device controls, volume, sleep timer and contextual actions. Disabled or unavailable actions should explain why they cannot run.

## Planned offline behavior

Authorized content may support downloads or managed caches according to the source’s policy. Downloaded and Cached are different states. Corrupt, expired or unauthorized local assets must not continue to be presented as valid offline content.

External-provider playback availability does not automatically authorize downloading or caching that provider’s media.

## Privacy controls

Private listening data should remain private by default. Planned controls include provider-specific enablement, private listening/session behavior, family aggregation consent, local-first recommendations and explicit external-search behavior.

## Support and diagnostics

When released, Music should expose understandable diagnostics for source, authorization, routing, transcoding, availability, provider degradation and recovery without exposing private listening content in ordinary operational telemetry.

Until a release is separately validated and published, repository documentation and planned workflows are not a statement of service availability.
