package sqlitestore

import (
	"context"
	"database/sql"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/storage"
	sqlite "modernc.org/sqlite"
)

func TestBackendMigratesAndPersistsLibraryAuthorization(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "music.db")
	backend, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	version, err := backend.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion() error = %v", err)
	}
	if version != storage.CurrentVersion {
		t.Fatalf("schema version = %d, want %d", version, storage.CurrentVersion)
	}

	alice := domain.ProfileID("profile:alice")
	bob := domain.ProfileID("profile:bob")
	library := domain.LibraryID("library:home")
	if err := backend.CreateProfile(ctx, alice, "Alice"); err != nil {
		t.Fatalf("CreateProfile(alice) error = %v", err)
	}
	if err := backend.CreateProfile(ctx, bob, "Bob"); err != nil {
		t.Fatalf("CreateProfile(bob) error = %v", err)
	}
	if err := backend.CreateLibrary(ctx, library, alice, "Home Library", filepath.Join(t.TempDir(), "media")); err != nil {
		t.Fatalf("CreateLibrary() error = %v", err)
	}

	assertPermission(t, ctx, backend, alice, library, LibraryPermissionOwner, true)
	assertPermission(t, ctx, backend, alice, library, LibraryPermissionRead, true)
	assertPermission(t, ctx, backend, bob, library, LibraryPermissionRead, false)

	if err := backend.SetLibraryPermission(ctx, library, bob, LibraryPermissionRead); err != nil {
		t.Fatalf("SetLibraryPermission() error = %v", err)
	}
	assertPermission(t, ctx, backend, bob, library, LibraryPermissionRead, true)
	assertPermission(t, ctx, backend, bob, library, LibraryPermissionEdit, false)

	bobLibraries, err := backend.LibrariesForProfile(ctx, bob)
	if err != nil {
		t.Fatalf("LibrariesForProfile() error = %v", err)
	}
	if len(bobLibraries) != 1 || bobLibraries[0].ID != library || bobLibraries[0].Permission != LibraryPermissionRead {
		t.Fatalf("bob libraries = %#v", bobLibraries)
	}

	if err := backend.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	backend, err = Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen error = %v", err)
	}
	defer backend.Close()
	assertPermission(t, ctx, backend, bob, library, LibraryPermissionRead, true)
}

func TestMigrationFromVersionOnePreservesOwnerAuthorization(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "music.db")
	dsn, err := DSN(path)
	if err != nil {
		t.Fatalf("DSN() error = %v", err)
	}
	connector, err := sqlite.NewConnector(dsn)
	if err != nil {
		t.Fatalf("NewConnector() error = %v", err)
	}
	db := sql.OpenDB(connector)
	legacy := &Backend{db: db}
	if err := storage.Migrate(ctx, legacy, 1); err != nil {
		t.Fatalf("migrate to schema v1: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(ctx, `INSERT INTO profiles(profile_id, display_name, created_at, updated_at) VALUES(?, ?, ?, ?)`, "profile:owner", "Owner", now, now); err != nil {
		t.Fatalf("insert v1 profile: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO libraries(library_id, owner_profile_id, name, root_path, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?)`, "library:legacy", "profile:owner", "Legacy", filepath.Join(t.TempDir(), "legacy-media"), now, now); err != nil {
		t.Fatalf("insert v1 library: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close v1 database: %v", err)
	}

	backend, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open() after v1 state error = %v", err)
	}
	defer backend.Close()
	assertPermission(t, ctx, backend, domain.ProfileID("profile:owner"), domain.LibraryID("library:legacy"), LibraryPermissionOwner, true)
	libraries, err := backend.LibrariesForProfile(ctx, domain.ProfileID("profile:owner"))
	if err != nil {
		t.Fatalf("LibrariesForProfile() after migration error = %v", err)
	}
	if len(libraries) != 1 || libraries[0].ID != domain.LibraryID("library:legacy") {
		t.Fatalf("migrated owner libraries = %#v", libraries)
	}
}

func TestLibraryAuthorizationFailsClosed(t *testing.T) {
	ctx := context.Background()
	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer backend.Close()

	alice := domain.ProfileID("profile:alice")
	bob := domain.ProfileID("profile:bob")
	library := domain.LibraryID("library:private")
	if err := backend.CreateProfile(ctx, alice, "Alice"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateProfile(ctx, bob, "Bob"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, alice, "Private", filepath.Join(t.TempDir(), "private")); err != nil {
		t.Fatal(err)
	}

	if err := backend.SetLibraryPermission(ctx, library, bob, LibraryPermissionOwner); err == nil {
		t.Fatal("expected ownership grant without transfer to fail")
	}
	if err := backend.SetLibraryPermission(ctx, library, alice, LibraryPermissionRead); err == nil {
		t.Fatal("expected owner downgrade to fail")
	}
	if err := backend.CreateLibrary(ctx, domain.LibraryID("library:relative"), alice, "Relative", "relative/path"); err == nil {
		t.Fatal("expected relative library path to fail")
	}
}

func TestDSNEnablesRequiredSafetyControls(t *testing.T) {
	dsn, err := DSN(filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatalf("DSN() error = %v", err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	want := map[string]string{
		"_busy_timeout": "5000",
		"_defensive":    "1",
		"_foreign_keys": "on",
		"_journal_mode": "WAL",
		"_synchronous":  "FULL",
	}
	for key, expected := range want {
		if got := parsed.Query().Get(key); got != expected {
			t.Fatalf("DSN setting %s = %q, want %q", key, got, expected)
		}
	}
}

func assertPermission(t *testing.T, ctx context.Context, backend *Backend, profile domain.ProfileID, library domain.LibraryID, permission LibraryPermission, want bool) {
	t.Helper()
	got, err := backend.HasLibraryPermission(ctx, profile, library, permission)
	if err != nil {
		t.Fatalf("HasLibraryPermission() error = %v", err)
	}
	if got != want {
		t.Fatalf("HasLibraryPermission(%s) = %v, want %v", permission, got, want)
	}
}
