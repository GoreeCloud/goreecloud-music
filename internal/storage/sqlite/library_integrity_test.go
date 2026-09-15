package sqlitestore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func TestHashLibraryFileContentPersistsStableSHA256Idempotently(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "Artist", "Album", "Track.mp3")
	media := append(testID3v23(map[string]string{
		"TIT2": "Track",
		"TPE1": "Artist",
		"TALB": "Album",
		"TPE2": "Album Artist",
	}), testIngestMP3Frames()...)
	mustWriteLibraryFile(t, track, media)

	backend, owner, reader, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	ingest, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:hash"), domain.ReleaseID("release:hash"), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	got, err := backend.HashLibraryFileContent(ctx, owner, library, fileID)
	if err != nil {
		t.Fatalf("HashLibraryFileContent() error = %v", err)
	}
	expected := sha256.Sum256(media)
	if got.PlayableAssetID != ingest.PlayableAssetID || got.SHA256 != hex.EncodeToString(expected[:]) || got.SizeBytes != int64(len(media)) {
		t.Fatalf("content hash = %#v", got)
	}
	again, err := backend.HashLibraryFileContent(ctx, owner, library, fileID)
	if err != nil || again != got {
		t.Fatalf("idempotent content hash = %#v, err = %v", again, err)
	}

	var stored string
	if err := backend.db.QueryRowContext(ctx, `SELECT media_sha256 FROM playable_assets WHERE playable_asset_id = ?`, string(got.PlayableAssetID)).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != got.SHA256 {
		t.Fatalf("stored hash = %q, want %q", stored, got.SHA256)
	}
	if _, err := backend.HashLibraryFileContent(ctx, reader, library, fileID); err == nil {
		t.Fatal("expected read-only profile hashing to fail")
	}
}

func TestHashLibraryFileContentRequiresMaterializedAsset(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mustWriteLibraryFile(t, track, testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}))
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	if _, err := backend.HashLibraryFileContent(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected hashing before materialization to fail")
	}
}

func TestHashLibraryFileContentRejectsChangedSourceWithoutMutation(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	media := testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"})
	mustWriteLibraryFile(t, track, media)
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	ingest, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:changed-hash"), domain.ReleaseID("release:changed-hash"), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	mustWriteLibraryFile(t, track, append(media, 0))
	if _, err := backend.HashLibraryFileContent(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected changed source hashing to fail")
	}
	var stored sql.NullString
	if err := backend.db.QueryRowContext(ctx, `SELECT media_sha256 FROM playable_assets WHERE playable_asset_id = ?`, string(ingest.PlayableAssetID)).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored.Valid {
		t.Fatalf("failed hash persisted state: %#v", stored)
	}
}

func TestHashLibraryFileContentRejectsConflictingStoredHash(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mustWriteLibraryFile(t, track, testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}))
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	ingest, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:hash-conflict"), domain.ReleaseID("release:hash-conflict"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	conflict := strings.Repeat("0", 64)
	if _, err := backend.db.ExecContext(ctx, `UPDATE playable_assets SET media_sha256 = ? WHERE playable_asset_id = ?`, conflict, string(ingest.PlayableAssetID)); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.HashLibraryFileContent(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected conflicting stored hash to fail closed")
	}
	var stored string
	if err := backend.db.QueryRowContext(ctx, `SELECT media_sha256 FROM playable_assets WHERE playable_asset_id = ?`, string(ingest.PlayableAssetID)).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != conflict {
		t.Fatalf("conflicting stored hash was overwritten: %q", stored)
	}
}

func TestHashLibraryFileContentMissingSourceFails(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mustWriteLibraryFile(t, track, testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}))
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	if _, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:missing-hash"), domain.ReleaseID("release:missing-hash"), time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(track); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.HashLibraryFileContent(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected missing source hashing to fail")
	}
}
