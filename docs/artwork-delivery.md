# Resonance artwork delivery

GoreeCloud Music serves album artwork only after the request principal is verified against persisted identity and authorized to read the requested library.

## Endpoint

`GET /api/v1/libraries/{libraryID}/albums/{albumID}/artwork`

The album lookup is scoped to the requested library. Ordinary users require a readable library membership; administrators follow the existing administrative library policy.

## Source handling

Sidecar artwork is streamed from the recorded local source file. Embedded artwork is extracted on demand with local FFmpeg from the recorded source audio. No source music file or sidecar image is rewritten, moved, retagged, or deleted.

Responses declare the stored image MIME type, use `X-Content-Type-Options: nosniff`, and are private-cacheable for a bounded period. External artwork providers are not required.

Thumbnail generation, derivative caching, conditional request validators, and optimized embedded-art extraction remain later performance work. The authoritative source continues to be the user-owned local media collection.
