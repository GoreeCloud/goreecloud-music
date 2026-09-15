package sqlitestore

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/libraryscan"
	"github.com/GoreeCloud/goreecloud-music/internal/storage"
	sqlite "modernc.org/sqlite"
)

func TestScanLibraryReconcilesFilesWithoutMutatingSources(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	trackA := filepath.Join(root, "Artist", "Alpha.flac")
	trackB := filepath.Join(root, "Beta.mp3")
	ignored := filepath.Join(root, "notes.txt")
	mustWriteLibraryFile(t, trackA, []byte("alpha"))
	mustWriteLibraryFile(t, trackB, []byte("beta"))
	mustWriteLibraryFile(t, ignored, []byte("ignore"))

	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	owner := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:scan")
	if err := backend.CreateProfile(ctx, owner, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Scan Library", root); err != nil {
		t.Fatal(err)
	}

	firstScan := time.Date(2026, 9, 15, 20, 20, 0, 0, time.UTC)
	result, err := backend.ScanLibrary(ctx, owner, library, firstScan)
	if err != nil {
		t.Fatalf("ScanLibrary() error = %v", err)
	}
	if result.Added != 2 || result.Updated != 0 || result.Missing != 0 || result.Restored != 0 {
		t.Fatalf("first result = %#v", result)
	}
	files, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].RelativePath != "Artist/Alpha.flac" || files[1].RelativePath != "Beta.mp3" {
		t.Fatalf("files = %#v", files)
	}
	alphaBefore, err := os.ReadFile(trackA)
	if err != nil {
		t.Fatal(err)
	}
	if string(alphaBefore) != "alpha" {
		t.Fatalf("source changed after scan: %q", alphaBefore)
	}

	if err := os.WriteFile(trackA, []byte("alpha-updated"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(trackB); err != nil {
		t.Fatal(err)
	}
	secondScan := firstScan.Add(time.Minute)
	result, err = backend.ScanLibrary(ctx, owner, library, secondScan)
	if err != nil {
		t.Fatalf("second ScanLibrary() error = %v", err)
	}
	if result.Updated != 1 || result.Missing != 1 || result.Added != 0 || result.Restored != 0 {
		t.Fatalf("second result = %#v", result)
	}
	visible, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(visible) != 1 || visible[0].RelativePath != "Artist/Alpha.flac" {
		t.Fatalf("visible files = %#v", visible)
	}
	withMissing, err := backend.LibraryFilesForProfile(ctx, owner, library, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(withMissing) != 2 || withMissing[1].MissingSince == nil {
		t.Fatalf("files with tombstone = %#v", withMissing)
	}

	mustWriteLibraryFile(t, trackB, []byte("beta-returned"))
	thirdScan := secondScan.Add(time.Minute)
	result, err = backend.ScanLibrary(ctx, owner, library, thirdScan)
	if err != nil {
		t.Fatalf("third ScanLibrary() error = %v", err)
	}
	if result.Restored != 1 || result.Missing != 0 {
		t.Fatalf("third result = %#v", result)
	}
	visible, err = backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(visible) != 2 {
		t.Fatalf("restored visible files = %#v", visible)
	}
}

func TestLibraryFileOperationsEnforceMembershipPermissions(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	mustWriteLibraryFile(t, filepath.Join(root, "track.ogg"), []byte("ogg"))
	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	owner := domain.ProfileID("profile:owner")
	reader := domain.ProfileID("profile:reader")
	outsider := domain.ProfileID("profile:outsider")
	library := domain.LibraryID("library:private")
	for _, profile := range []domain.ProfileID{owner, reader, outsider} {
		if err := backend.CreateProfile(ctx, profile, string(profile)); err != nil {
			t.Fatal(err)
		}
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Private", root); err != nil {
		t.Fatal(err)
	}
	if err := backend.SetLibraryPermission(ctx, library, reader, LibraryPermissionRead); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ScanLibrary(ctx, owner, library, time.Now()); err != nil {
		t.Fatal(err)
	}

	if files, err := backend.LibraryFilesForProfile(ctx, reader, library, false); err != nil || len(files) != 1 {
		t.Fatalf("reader list files = %#v, err = %v", files, err)
	}
	if _, err := backend.ScanLibrary(ctx, reader, library, time.Now()); err == nil {
		t.Fatal("expected read-only profile scan to fail")
	}
	if _, err := backend.LibraryFilesForProfile(ctx, outsider, library, false); err == nil {
		t.Fatal("expected unauthorized profile listing to fail")
	}
	if _, err := backend.ReconcileLibraryFiles(ctx, outsider, library, []libraryscan.Observation{{RelativePath: "track.ogg", SizeBytes: 3}}, time.Now()); err == nil {
		t.Fatal("expected unauthorized reconciliation to fail")
	}
}

func TestSchemaV2MigratesToLibraryFilesWithoutChangingAuthorization(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "music.db")
	dsn, err := DSN(path)
	if err != nil {
		t.Fatal(err)
	}
	connector, err := sqlite.NewConnector(dsn)
	if err != nil {
		t.Fatal(err)
	}
	db := sql.OpenDB(connector)
	legacy := &Backend{db: db}
	if err := storage.Migrate(ctx, legacy, 2); err != nil {
		t.Fatalf("migrate to v2: %v", err)
	}
	owner := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:legacy-v2")
	if err := legacy.CreateProfile(ctx, owner, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := legacy.CreateLibrary(ctx, library, owner, "Legacy v2", t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}

	backend, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	version, err := backend.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != storage.CurrentVersion {
		t.Fatalf("version = %d want %d", version, storage.CurrentVersion)
	}
	allowed, err := backend.HasLibraryPermission(ctx, owner, library, LibraryPermissionOwner)
	if err != nil || !allowed {
		t.Fatalf("owner authorization after v3 migration = %v, err = %v", allowed, err)
	}
}

func TestReconcileRejectsEscapingAndDuplicatePaths(t *testing.T) {
	ctx := context.Background()
	backend, owner, library := testLibraryBackend(t, ctx)
	defer backend.Close()
	if _, err := backend.ReconcileLibraryFiles(ctx, owner, library, []libraryscan.Observation{{RelativePath: "../escape.mp3"}}, time.Now()); err == nil {
		t.Fatal("expected escaping path to fail")
	}
	if _, err := backend.ReconcileLibraryFiles(ctx, owner, library, []libraryscan.Observation{{RelativePath: "same.mp3"}, {RelativePath: "same.mp3"}}, time.Now()); err == nil {
		t.Fatal("expected duplicate path to fail")
	}
}

func testLibraryBackend(t *testing.T, ctx context.Context) (*Backend, domain.ProfileID, domain.LibraryID) {
	t.Helper()
	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	owner := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:test")
	if err := backend.CreateProfile(ctx, owner, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Test", t.TempDir()); err != nil {
		t.Fatal(err)
	}
	return backend, owner, library
}

func mustWriteLibraryFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}
