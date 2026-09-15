# GoreeCloud Music — Security

## Status

GoreeCloud Music is in Active Development and is not production-ready or Stable.

## Current boundary

Milestone 0 implements architecture contracts and a local development service only. It does not implement production authentication, external-provider credentials, music-library access, downloads, or public deployment.

The default bind is loopback-only (`127.0.0.1:8080`) to avoid accidental network exposure during foundation development.

## Reporting

Do not post credentials, private user data, provider tokens, private infrastructure details, or exploit-sensitive operational evidence in public issues. Security reports should use an approved private GoreeCloud reporting path when available.

## Required future controls

Before production eligibility, the application must provide verified authentication/authorization, multi-user isolation, Privacy Shield and Wardveil integration where applicable, abuse/resource controls, safe secret handling, dependency review, secure deployment defaults, audit-safe logging, backup/recovery protection, and security testing.
