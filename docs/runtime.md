# Runtime and database lifecycle

GoreeCloud Music keeps database startup explicit and bounded.

## PostgreSQL

`DATABASE_URL` enables PostgreSQL-backed runtime behavior. The server uses the pgx `database/sql` driver, validates connectivity at startup, applies bounded connection-pool settings, and closes the pool during process shutdown.

If `DATABASE_URL` is absent, the service can still expose non-persistent foundation endpoints, but persistence-backed endpoints are unavailable.

## Migrations

Versioned SQL migrations are embedded from `migrations/*.sql`.

Automatic startup migration is disabled by default. Set:

```text
GOREECLOUD_MUSIC_AUTO_MIGRATE=true
```

only in an environment where the operator has explicitly approved schema changes. Applied versions are recorded in `schema_migrations` and each migration is executed transactionally.

## Development identity

`GOREECLOUD_MUSIC_DEV_IDENTITY=true` enables the development-only header identity middleware. This exists only to exercise user-scoped APIs before GoreeCloud Identity integration and must not be enabled as a substitute for production authentication.

## Shutdown

SIGINT and SIGTERM trigger bounded HTTP graceful shutdown. Database connections are closed after the server exits.
