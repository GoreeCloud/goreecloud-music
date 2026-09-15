package sqlitestore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func TestHashLibraryFileContentPersistsExactSourceDigestIdempotently(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "Artist", "Album", "Track.mp3")
	mediaBytes := append(testID3v23(map[string]string{
		"TIT2": "Track",
		"TPE1": "Artist",
		"TALB": "Album",
		"TPE2": "Album Artist",
	}), testIngestMP3Frames()...)
	mustWriteLibraryFile(t, track, mediaBytes)

	backend, owner, reader, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	ingest, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:integrity"), domain.ReleaseID("release:integrity"), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	got, err := backend.HashLibraryFileContent(ctx, owner, library, fileID)
	if err != nil {
		t.Fatalf("HashLibraryFileContent() error = %v", err)
	}
	expected := sha256.Sum256(mediaBytes)
	expectedHex := hex.EncodeToString(expected[:])
	if got.PlayableAssetID != ingest.PlayableAssetID || got.SHA256 != expectedHex {
		t.Fatalf("integrity = %#v, expected hash %q", got, expectedHex)
	}

	again, err := backend.HashLibraryFileContent(ctx, owner, library, fileID)
	if err != nil || again != got {
		t.Fatalf("idempotent hash = %#v, err = %v", again, err)
	}

	var storedHash string
	var codec, container sql.NullString
	if err := backend.db.QueryRowContext(ctx, `SELECT media_sha256, codec, container FROM playable_assets WHERE playable_asset_id = ?`, string(got.PlayableAssetID)).Scan(&storedHash, &codec, &container); err != nil {
		t.Fatal(err)
	}
	if storedHash != expectedHex || codec.Valid || container.Valid {
		t.Fatalf("stored integrity = hash=%q codec=%#v container=%#v", storedHash, codec, container)
	}

	if _, err := backend.HashLibraryFileContent(ctx, reader, library, fileID); err == nil {
		t.Fatal("expected read-only profile integrity hashing to fail")
	}
}

func TestHashLibraryFileContentChangedSourcePreservesExistingHash(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mediaBytes := append(testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}), testIngestMP3Frames()...)
	mustWriteLibraryFile(t, track, mediaBytes)

	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	ingest, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:integrity-stale"), domain.ReleaseID("release:integrity-stale"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	first, err := backend.HashLibraryFileContent(ctx, owner, library, fileID)
	if err != nil {
		t.Fatal(err)
	}

	mustWriteLibraryFile(t, track, append(mediaBytes, 0))
	if _, err := backend.HashLibraryFileContent(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected integrity hashing changed source to fail")
	}

	var storedHash string
	if err := backend.db.QueryRowContext(ctx, `SELECT media_sha256 FROM playable_assets WHERE playable_asset_id = ?`, string(ingest.PlayableAssetID)).Scan(&storedHash); err != nil {
		t.Fatal(err)
	}
	if storedHash != first.SHA256 {
		t.Fatalf("failed integrity refresh replaced prior hash: got %q want %q", storedHash, first.SHA256)
	}
}

func TestHashLibraryFileContentRequiresMaterializedAsset(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mustWriteLibraryFile(t, track, append(testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}), testIngestMP3Frames()...))
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()

	if _, err := backend.HashLibraryFileContent(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected integrity hashing before materialization to fail")
	}
}

func TestSHA256WithContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := sha256WithContext(ctx, strings.NewReader("goreecloud music")); err == nil {
		t.Fatal("expected canceled hashing context to fail")
	}
}
