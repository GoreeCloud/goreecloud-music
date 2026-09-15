package sqlitestore

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/storage"
	sqlite "modernc.org/sqlite"
)

func TestMigrationFromVersionThreePreservesLibraryFiles(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	if err := os.WriteFile(track, testID3v23(map[string]string{"TIT2": "Before Migration"}), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

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
	if err := storage.Migrate(ctx, legacy, 3); err != nil {
		t.Fatalf("migrate to schema v3: %v", err)
	}

	owner := domain.ProfileID("profile:migration-owner")
	library := domain.LibraryID("library:migration-v3")
	if err := legacy.CreateProfile(ctx, owner, "Migration Owner"); err != nil {
		t.Fatalf("CreateProfile() at schema v3: %v", err)
	}
	if err := legacy.CreateLibrary(ctx, library, owner, "Migration Library", root); err != nil {
		t.Fatalf("CreateLibrary() at schema v3: %v", err)
	}
	scanAt := time.Date(2026, 9, 15, 22, 20, 0, 0, time.UTC)
	if _, err := legacy.ScanLibrary(ctx, owner, library, scanAt); err != nil {
		t.Fatalf("ScanLibrary() at schema v3: %v", err)
	}
	before, err := legacy.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(before) != 1 {
		t.Fatalf("schema-v3 files = %#v, err = %v", before, err)
	}
	beforeID := before[0].ID
	if err := db.Close(); err != nil {
		t.Fatalf("close schema-v3 database: %v", err)
	}

	backend, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open() after schema-v3 state: %v", err)
	}
	defer backend.Close()
	version, err := backend.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion() after migration: %v", err)
	}
	if version != storage.CurrentVersion {
		t.Fatalf("schema version = %d, want %d", version, storage.CurrentVersion)
	}
	after, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(after) != 1 {
		t.Fatalf("schema-v4 files = %#v, err = %v", after, err)
	}
	if after[0].ID != beforeID || after[0].RelativePath != "track.mp3" {
		t.Fatalf("migrated file = %#v, want id %q path track.mp3", after[0], beforeID)
	}
	if _, found, err := backend.LibraryFileMetadataForProfile(ctx, owner, library, beforeID); err != nil || found {
		t.Fatalf("metadata before extraction found = %v, err = %v", found, err)
	}
}
