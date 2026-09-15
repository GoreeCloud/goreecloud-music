package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/storage"
	sqlite "modernc.org/sqlite"
)

// Backend is the GoreeCloud Music SQLite application-state backend. It stores
// metadata and authorization state only; original media bytes remain in the
// separately managed library filesystem.
type Backend struct {
	db *sql.DB
}

// LibraryPermission is the explicit per-profile authorization level for a
// library. Service-layer authorization must still validate the requesting
// profile and purpose before invoking mutations.
type LibraryPermission string

const (
	LibraryPermissionRead  LibraryPermission = "read"
	LibraryPermissionEdit  LibraryPermission = "edit"
	LibraryPermissionOwner LibraryPermission = "owner"
)

// Library is durable library identity and source-root metadata.
type Library struct {
	ID             domain.LibraryID
	OwnerProfileID domain.ProfileID
	Name           string
	RootPath       string
	Permission     LibraryPermission
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

var entityDDL = map[string]string{
	"schema_metadata": `CREATE TABLE IF NOT EXISTS schema_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	) STRICT`,
	"profiles": `CREATE TABLE IF NOT EXISTS profiles (
		profile_id TEXT PRIMARY KEY,
		display_name TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	) STRICT`,
	"libraries": `CREATE TABLE IF NOT EXISTS libraries (
		library_id TEXT PRIMARY KEY,
		owner_profile_id TEXT NOT NULL REFERENCES profiles(profile_id) ON DELETE RESTRICT,
		name TEXT NOT NULL,
		root_path TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		UNIQUE(owner_profile_id, root_path)
	) STRICT`,
	"library_memberships": `CREATE TABLE IF NOT EXISTS library_memberships (
		library_id TEXT NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
		profile_id TEXT NOT NULL REFERENCES profiles(profile_id) ON DELETE CASCADE,
		permission TEXT NOT NULL CHECK(permission IN ('read', 'edit', 'owner')),
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		PRIMARY KEY(library_id, profile_id)
	) STRICT`,
	"recordings": `CREATE TABLE IF NOT EXISTS recordings (
		recording_id TEXT PRIMARY KEY,
		library_id TEXT NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		artist TEXT NOT NULL DEFAULT '',
		release_id TEXT,
		metadata_json TEXT NOT NULL DEFAULT '{}',
		added_at TEXT NOT NULL
	) STRICT`,
	"releases": `CREATE TABLE IF NOT EXISTS releases (
		release_id TEXT PRIMARY KEY,
		library_id TEXT NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		artist TEXT NOT NULL DEFAULT '',
		release_date TEXT,
		metadata_json TEXT NOT NULL DEFAULT '{}'
	) STRICT`,
	"source_items": `CREATE TABLE IF NOT EXISTS source_items (
		source_item_id TEXT PRIMARY KEY,
		library_id TEXT NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
		recording_id TEXT REFERENCES recordings(recording_id) ON DELETE CASCADE,
		source_kind TEXT NOT NULL,
		external_id TEXT,
		metadata_json TEXT NOT NULL DEFAULT '{}'
	) STRICT`,
	"playable_assets": `CREATE TABLE IF NOT EXISTS playable_assets (
		playable_asset_id TEXT PRIMARY KEY,
		library_id TEXT NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
		recording_id TEXT NOT NULL REFERENCES recordings(recording_id) ON DELETE CASCADE,
		source_item_id TEXT REFERENCES source_items(source_item_id) ON DELETE SET NULL,
		media_path TEXT NOT NULL,
		media_sha256 TEXT,
		media_size_bytes INTEGER,
		codec TEXT,
		container TEXT,
		duration_ms INTEGER,
		added_at TEXT NOT NULL
	) STRICT`,
	"queues": `CREATE TABLE IF NOT EXISTS queues (
		queue_id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL REFERENCES profiles(profile_id) ON DELETE CASCADE,
		revision INTEGER NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL
	) STRICT`,
	"queue_items": `CREATE TABLE IF NOT EXISTS queue_items (
		queue_item_id TEXT PRIMARY KEY,
		queue_id TEXT NOT NULL REFERENCES queues(queue_id) ON DELETE CASCADE,
		position INTEGER NOT NULL,
		recording_id TEXT NOT NULL REFERENCES recordings(recording_id) ON DELETE RESTRICT,
		source_item_id TEXT REFERENCES source_items(source_item_id) ON DELETE SET NULL,
		selected_route TEXT,
		route_reason TEXT,
		UNIQUE(queue_id, position)
	) STRICT`,
	"favorites": `CREATE TABLE IF NOT EXISTS favorites (
		favorite_id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL REFERENCES profiles(profile_id) ON DELETE CASCADE,
		recording_id TEXT NOT NULL REFERENCES recordings(recording_id) ON DELETE CASCADE,
		created_at TEXT NOT NULL,
		UNIQUE(profile_id, recording_id)
	) STRICT`,
	"ratings": `CREATE TABLE IF NOT EXISTS ratings (
		rating_id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL REFERENCES profiles(profile_id) ON DELETE CASCADE,
		recording_id TEXT NOT NULL REFERENCES recordings(recording_id) ON DELETE CASCADE,
		rating INTEGER NOT NULL CHECK(rating >= 0 AND rating <= 100),
		updated_at TEXT NOT NULL,
		UNIQUE(profile_id, recording_id)
	) STRICT`,
	"play_history": `CREATE TABLE IF NOT EXISTS play_history (
		history_id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL REFERENCES profiles(profile_id) ON DELETE CASCADE,
		recording_id TEXT NOT NULL REFERENCES recordings(recording_id) ON DELETE CASCADE,
		played_at TEXT NOT NULL,
		played_ms INTEGER NOT NULL DEFAULT 0
	) STRICT`,
}

// Open opens, configures, migrates, and verifies a SQLite application-state
// database. The DSN enables foreign keys, WAL journaling, full synchronous
// durability, a bounded busy timeout, and SQLite defensive mode on every
// physical connection.
func Open(ctx context.Context, path string) (*Backend, error) {
	dsn, err := DSN(path)
	if err != nil {
		return nil, err
	}
	connector, err := sqlite.NewConnector(dsn)
	if err != nil {
		return nil, fmt.Errorf("create sqlite connector: %w", err)
	}
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	db.SetConnMaxIdleTime(5 * time.Minute)

	backend := &Backend{db: db}
	if err := backend.Ping(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := storage.Migrate(ctx, backend, storage.CurrentVersion); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate sqlite application state: %w", err)
	}
	return backend, nil
}

// DSN returns the controlled file URI used by the SQLite driver.
func DSN(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("sqlite database path must not be empty")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve sqlite database path: %w", err)
	}
	u := &url.URL{Scheme: "file", Path: filepath.ToSlash(absolute)}
	q := u.Query()
	q.Set("_busy_timeout", "5000")
	q.Set("_defensive", "1")
	q.Set("_foreign_keys", "on")
	q.Set("_journal_mode", "WAL")
	q.Set("_synchronous", "FULL")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (b *Backend) Close() error {
	if b == nil || b.db == nil {
		return nil
	}
	return b.db.Close()
}

func (b *Backend) Ping(ctx context.Context) error {
	if b == nil || b.db == nil {
		return fmt.Errorf("sqlite backend is not initialized")
	}
	if err := b.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sqlite application state: %w", err)
	}
	return nil
}

func (b *Backend) SchemaVersion(ctx context.Context) (storage.Version, error) {
	var exists int
	if err := b.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_metadata'`).Scan(&exists); err != nil {
		return storage.UninitializedVersion, err
	}
	if exists == 0 {
		return storage.UninitializedVersion, nil
	}
	var raw string
	err := b.db.QueryRowContext(ctx, `SELECT value FROM schema_metadata WHERE key = 'schema_version'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.UninitializedVersion, nil
	}
	if err != nil {
		return storage.UninitializedVersion, err
	}
	parsed, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return storage.UninitializedVersion, fmt.Errorf("parse schema version %q: %w", raw, err)
	}
	return storage.Version(parsed), nil
}

func (b *Backend) ApplyMigration(ctx context.Context, migration storage.Migration) error {
	current, err := b.SchemaVersion(ctx)
	if err != nil {
		return err
	}
	if current != migration.From {
		return fmt.Errorf("migration %s starts at %d, backend is %d", migration.ID, migration.From, current)
	}

	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, change := range migration.Changes {
		if change.Kind != storage.ChangeCreateEntity {
			return fmt.Errorf("migration %s contains unsupported change kind %q", migration.ID, change.Kind)
		}
		statement, ok := entityDDL[change.Entity]
		if !ok {
			return fmt.Errorf("migration %s has no sqlite mapping for entity %q", migration.ID, change.Entity)
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("create entity %s: %w", change.Entity, err)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	metadata := [][2]string{
		{"schema_version", strconv.FormatUint(uint64(migration.To), 10)},
		{"last_migration_id", migration.ID},
		{"last_migration_at", now},
	}
	for _, item := range metadata {
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_metadata(key, value) VALUES(?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, item[0], item[1]); err != nil {
			return fmt.Errorf("record migration metadata %s: %w", item[0], err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.ID, err)
	}
	return nil
}

func (b *Backend) CreateProfile(ctx context.Context, profileID domain.ProfileID, displayName string) error {
	validated, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return err
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return fmt.Errorf("profile display name must not be empty")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = b.db.ExecContext(ctx, `INSERT INTO profiles(profile_id, display_name, created_at, updated_at) VALUES(?, ?, ?, ?)`, string(validated), displayName, now, now)
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	return nil
}

func (b *Backend) CreateLibrary(ctx context.Context, libraryID domain.LibraryID, ownerProfileID domain.ProfileID, name, rootPath string) error {
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return err
	}
	validatedOwner, err := domain.NewProfileID(string(ownerProfileID))
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("library name must not be empty")
	}
	rootPath = strings.TrimSpace(rootPath)
	if rootPath == "" {
		return fmt.Errorf("library root path must not be empty")
	}
	if !filepath.IsAbs(rootPath) {
		return fmt.Errorf("library root path must be absolute")
	}
	rootPath = filepath.Clean(rootPath)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO libraries(library_id, owner_profile_id, name, root_path, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?)`, string(validatedLibrary), string(validatedOwner), name, rootPath, now, now); err != nil {
		return fmt.Errorf("create library: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO library_memberships(library_id, profile_id, permission, created_at, updated_at) VALUES(?, ?, 'owner', ?, ?)`, string(validatedLibrary), string(validatedOwner), now, now); err != nil {
		return fmt.Errorf("create owner library membership: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit library creation: %w", err)
	}
	return nil
}

func (b *Backend) SetLibraryPermission(ctx context.Context, libraryID domain.LibraryID, profileID domain.ProfileID, permission LibraryPermission) error {
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return err
	}
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return err
	}
	if err := validatePermission(permission); err != nil {
		return err
	}

	var owner string
	if err := b.db.QueryRowContext(ctx, `SELECT owner_profile_id FROM libraries WHERE library_id = ?`, string(validatedLibrary)).Scan(&owner); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("library not found")
		}
		return err
	}
	if owner == string(validatedProfile) && permission != LibraryPermissionOwner {
		return fmt.Errorf("library owner permission cannot be downgraded")
	}
	if owner != string(validatedProfile) && permission == LibraryPermissionOwner {
		return fmt.Errorf("owner permission cannot be granted without transferring library ownership")
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = b.db.ExecContext(ctx, `INSERT INTO library_memberships(library_id, profile_id, permission, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(library_id, profile_id) DO UPDATE SET permission = excluded.permission, updated_at = excluded.updated_at`, string(validatedLibrary), string(validatedProfile), string(permission), now, now)
	if err != nil {
		return fmt.Errorf("set library permission: %w", err)
	}
	return nil
}

func (b *Backend) HasLibraryPermission(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, required LibraryPermission) (bool, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return false, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return false, err
	}
	if err := validatePermission(required); err != nil {
		return false, err
	}
	var actual string
	err = b.db.QueryRowContext(ctx, `SELECT permission FROM library_memberships WHERE library_id = ? AND profile_id = ?`, string(validatedLibrary), string(validatedProfile)).Scan(&actual)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return permissionRank(LibraryPermission(actual)) >= permissionRank(required), nil
}

func (b *Backend) LibrariesForProfile(ctx context.Context, profileID domain.ProfileID) ([]Library, error) {
	validated, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return nil, err
	}
	rows, err := b.db.QueryContext(ctx, `SELECT l.library_id, l.owner_profile_id, l.name, l.root_path, m.permission, l.created_at, l.updated_at
		FROM library_memberships m JOIN libraries l ON l.library_id = m.library_id
		WHERE m.profile_id = ? ORDER BY lower(l.name), l.library_id`, string(validated))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libraries []Library
	for rows.Next() {
		var libraryID, ownerID, name, rootPath, permission, createdRaw, updatedRaw string
		if err := rows.Scan(&libraryID, &ownerID, &name, &rootPath, &permission, &createdRaw, &updatedRaw); err != nil {
			return nil, err
		}
		createdAt, err := time.Parse(time.RFC3339Nano, createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse library created_at: %w", err)
		}
		updatedAt, err := time.Parse(time.RFC3339Nano, updatedRaw)
		if err != nil {
			return nil, fmt.Errorf("parse library updated_at: %w", err)
		}
		libraries = append(libraries, Library{
			ID:             domain.LibraryID(libraryID),
			OwnerProfileID: domain.ProfileID(ownerID),
			Name:           name,
			RootPath:       rootPath,
			Permission:     LibraryPermission(permission),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return libraries, nil
}

func validatePermission(permission LibraryPermission) error {
	switch permission {
	case LibraryPermissionRead, LibraryPermissionEdit, LibraryPermissionOwner:
		return nil
	default:
		return fmt.Errorf("unsupported library permission %q", permission)
	}
}

func permissionRank(permission LibraryPermission) int {
	switch permission {
	case LibraryPermissionRead:
		return 1
	case LibraryPermissionEdit:
		return 2
	case LibraryPermissionOwner:
		return 3
	default:
		return 0
	}
}
