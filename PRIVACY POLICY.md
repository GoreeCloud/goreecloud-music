# GoreeCloud Music — Privacy Policy (Development State)

## Current implementation

Milestone 0 does not collect a music library, user accounts, listening history, external-provider credentials, analytics, telemetry, or advertising identifiers. The development service exposes only process health and build identity endpoints.

## Governing direction

GoreeCloud Music is designed for privacy by default. Future implementations must minimize data collection, keep listening state user-scoped by default, separate operational telemetry from listening history, and use Privacy Shield for applicable information-use decisions rather than inventing a parallel privacy authority.

## External providers

No external provider is enabled in the current implementation. Any future provider integration must be individually controllable and must document the data sent to that provider. Proposed YouTube support remains disabled and approval-required.

## Logs

The current service logs process startup and fatal server errors. Future content-level diagnostics must avoid exposing raw private listening metadata unless narrow authorized troubleshooting requires it.
