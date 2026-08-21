# Milestone 1 runtime checkpoint

This checkpoint advances Resonance Library infrastructure without authorizing production deployment.

Implemented in this branch:

- PostgreSQL pgx runtime driver wiring.
- Bounded connection-pool lifecycle and startup connectivity validation.
- Embedded SQL migrations with a schema migration ledger.
- Transactional migration execution gated behind an explicit opt-in environment variable.
- Persistence-backed router dependency wiring from the server runtime.
- Development-only identity wrapping from the runtime configuration boundary.
- Graceful SIGINT/SIGTERM HTTP shutdown.

Still pending before Milestone 1 is complete:

- GoreeCloud Identity production authentication.
- Scanner reconciliation into database records.
- Metadata and artwork persistence ingestion.
- Remaining library-management APIs and authorization coverage.
- Production deployment validation.

HTTP range streaming remains a Milestone 2 concern and is intentionally not introduced by this checkpoint.
