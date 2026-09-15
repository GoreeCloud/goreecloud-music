package sqlitestore

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func TestIngestLibraryFileCreatesCanonicalAndSourceStateIdempotently(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "Artist", "Album", "Track.mp3")
	mustWriteLibraryFile(t, track, testID3v23(map[string]string{
		"TIT2": "Track",
		"TPE1": "Artist",
		"TALB": "Album",
		"TPE2": "Album Artist",
	}))

	backend, owner, reader, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()

	when := time.Date(2026, 9, 15, 22, 40, 0, 0, time.UTC)
	got, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:track"), domain.ReleaseID("release:album"), when)
	if err != nil {
		t.Fatalf("IngestLibraryFile() error = %v", err)
	}
	if !got.CreatedRecording || !got.CreatedRelease || !got.CreatedSourceItem || !got.CreatedPlayableAsset {
		t.Fatalf("first ingest result = %#v", got)
	}
	if got.SourceItemID == "" || got.PlayableAssetID == "" {
		t.Fatalf("derived source identities missing: %#v", got)
	}

	again, err := backend.IngestLibraryFile(ctx, owner, library, fileID, got.RecordingID, got.ReleaseID, when.Add(time.Minute))
	if err != nil {
		t.Fatalf("idempotent IngestLibraryFile() error = %v", err)
	}
	if again.SourceItemID != got.SourceItemID || again.PlayableAssetID != got.PlayableAssetID {
		t.Fatalf("source identity changed across idempotent ingest: first=%#v second=%#v", got, again)
	}
	if again.CreatedRecording || again.CreatedRelease || again.CreatedSourceItem || again.CreatedPlayableAsset {
		t.Fatalf("second ingest should reuse existing state: %#v", again)
	}

	var releaseTitle, releaseArtist string
	if err := backend.db.QueryRowContext(ctx, `SELECT title, artist FROM releases WHERE release_id = ?`, string(got.ReleaseID)).Scan(&releaseTitle, &releaseArtist); err != nil {
		t.Fatal(err)
	}
	if releaseTitle != "Album" || releaseArtist != "Album Artist" {
		t.Fatalf("release = %q / %q", releaseTitle, releaseArtist)
	}

	var recordingTitle, recordingArtist, recordingRelease string
	if err := backend.db.QueryRowContext(ctx, `SELECT title, artist, release_id FROM recordings WHERE recording_id = ?`, string(got.RecordingID)).Scan(&recordingTitle, &recordingArtist, &recordingRelease); err != nil {
		t.Fatal(err)
	}
	if recordingTitle != "Track" || recordingArtist != "Artist" || recordingRelease != string(got.ReleaseID) {
		t.Fatalf("recording = %q / %q / %q", recordingTitle, recordingArtist, recordingRelease)
	}

	var sourceKind, externalID string
	if err := backend.db.QueryRowContext(ctx, `SELECT source_kind, external_id FROM source_items WHERE source_item_id = ?`, string(got.SourceItemID)).Scan(&sourceKind, &externalID); err != nil {
		t.Fatal(err)
	}
	if sourceKind != string(domain.SourceGoreeCloudServer) || externalID != fileID {
		t.Fatalf("source item = %q / %q", sourceKind, externalID)
	}

	var mediaPath string
	var mediaSize int64
	var codec, container sql.NullString
	if err := backend.db.QueryRowContext(ctx, `SELECT media_path, media_size_bytes, codec, container FROM playable_assets WHERE playable_asset_id = ?`, string(got.PlayableAssetID)).Scan(&mediaPath, &mediaSize, &codec, &container); err != nil {
		t.Fatal(err)
	}
	if mediaPath != "Artist/Album/Track.mp3" || mediaSize <= 0 || codec.Valid || container.Valid {
		t.Fatalf("playable asset = path=%q size=%d codec=%#v container=%#v", mediaPath, mediaSize, codec, container)
	}

	if _, err := backend.IngestLibraryFile(ctx, reader, library, fileID, domain.RecordingID("recording:reader"), domain.ReleaseID("release:reader"), when); err == nil {
		t.Fatal("expected read-only profile ingestion to fail")
	}
}

func TestIngestLibraryFileRequiresCurrentMetadataAndFilesystemSnapshot(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mustWriteLibraryFile(t, track, testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}))

	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	owner := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:ingest-current")
	if err := backend.CreateProfile(ctx, owner, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Library", root); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ScanLibrary(ctx, owner, library, time.Now()); err != nil {
		t.Fatal(err)
	}
	files, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(files) != 1 {
		t.Fatalf("files = %#v, err = %v", files, err)
	}
	fileID := files[0].ID

	if _, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:no-metadata"), domain.ReleaseID("release:no-metadata"), time.Now()); err == nil {
		t.Fatal("expected ingestion without extracted metadata to fail")
	}
	if _, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, fileID, time.Now()); err != nil {
		t.Fatal(err)
	}

	mustWriteLibraryFile(t, track, testID3v23(map[string]string{"TIT2": "Changed Title With New Size", "TALB": "Album"}))
	if _, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:stale"), domain.ReleaseID("release:stale"), time.Now()); err == nil {
		t.Fatal("expected ingestion against changed source file to fail")
	}
}

func TestIngestLibraryFileConflictRollsBackTransaction(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	firstPath := filepath.Join(root, "first.mp3")
	secondPath := filepath.Join(root, "second.mp3")
	mustWriteLibraryFile(t, firstPath, testID3v23(map[string]string{"TIT2": "First", "TALB": "Album"}))
	mustWriteLibraryFile(t, secondPath, testID3v23(map[string]string{"TIT2": "Second", "TALB": "Other Album"}))

	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	owner := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:ingest-conflict")
	if err := backend.CreateProfile(ctx, owner, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Library", root); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ScanLibrary(ctx, owner, library, time.Now()); err != nil {
		t.Fatal(err)
	}
	files, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(files) != 2 {
		t.Fatalf("files = %#v, err = %v", files, err)
	}
	byPath := map[string]string{}
	for _, file := range files {
		byPath[file.RelativePath] = file.ID
		if _, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, file.ID, time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := backend.IngestLibraryFile(ctx, owner, library, byPath["first.mp3"], domain.RecordingID("recording:shared"), domain.ReleaseID("release:first"), time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.IngestLibraryFile(ctx, owner, library, byPath["second.mp3"], domain.RecordingID("recording:shared"), domain.ReleaseID("release:should-rollback"), time.Now()); err == nil {
		t.Fatal("expected conflicting canonical recording id to fail")
	}

	var releaseCount int
	if err := backend.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM releases WHERE release_id = 'release:should-rollback'`).Scan(&releaseCount); err != nil {
		t.Fatal(err)
	}
	if releaseCount != 0 {
		t.Fatalf("conflicting ingestion left release state behind: count=%d", releaseCount)
	}
	var sourceCount int
	if err := backend.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM source_items WHERE external_id = ?`, byPath["second.mp3"]).Scan(&sourceCount); err != nil {
		t.Fatal(err)
	}
	if sourceCount != 0 {
		t.Fatalf("conflicting ingestion left source state behind: count=%d", sourceCount)
	}
}

func TestIngestLibraryFileRejectsIncompleteMetadataAndInvalidIDs(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	mustWriteLibraryFile(t, filepath.Join(root, "track.mp3"), testID3v23(map[string]string{"TIT2": "Track"}))

	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	owner := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:ingest-validation")
	if err := backend.CreateProfile(ctx, owner, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Library", root); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ScanLibrary(ctx, owner, library, time.Now()); err != nil {
		t.Fatal(err)
	}
	files, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(files) != 1 {
		t.Fatalf("files = %#v, err = %v", files, err)
	}
	if _, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, files[0].ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.IngestLibraryFile(ctx, owner, library, files[0].ID, domain.RecordingID("recording:track"), domain.ReleaseID("release:album"), time.Now()); err == nil {
		t.Fatal("expected album-less metadata ingestion to fail")
	}
	if _, err := backend.IngestLibraryFile(ctx, owner, library, files[0].ID, domain.RecordingID("recording/bad"), domain.ReleaseID("release:album"), time.Now()); err == nil {
		t.Fatal("expected invalid recording id to fail")
	}
}

func prepareIngestFixture(t *testing.T, ctx context.Context, root string) (*Backend, domain.ProfileID, domain.ProfileID, domain.LibraryID, string) {
	t.Helper()
	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	owner := domain.ProfileID("profile:owner")
	reader := domain.ProfileID("profile:reader")
	library := domain.LibraryID("library:ingest")
	for _, profile := range []domain.ProfileID{owner, reader} {
		if err := backend.CreateProfile(ctx, profile, string(profile)); err != nil {
			backend.Close()
			t.Fatal(err)
		}
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Library", root); err != nil {
		backend.Close()
		t.Fatal(err)
	}
	if err := backend.SetLibraryPermission(ctx, library, reader, LibraryPermissionRead); err != nil {
		backend.Close()
		t.Fatal(err)
	}
	if _, err := backend.ScanLibrary(ctx, owner, library, time.Now()); err != nil {
		backend.Close()
		t.Fatal(err)
	}
	files, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(files) != 1 {
		backend.Close()
		t.Fatalf("files = %#v, err = %v", files, err)
	}
	if _, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, files[0].ID, time.Now()); err != nil {
		backend.Close()
		t.Fatal(err)
	}
	return backend, owner, reader, library, files[0].ID
}

func TestIngestLibraryFileMissingSourceFailsWithoutState(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mustWriteLibraryFile(t, track, testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}))
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	if err := os.Remove(track); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:missing"), domain.ReleaseID("release:missing"), time.Now()); err == nil {
		t.Fatal("expected missing source ingestion to fail")
	}
	var count int
	if err := backend.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM recordings WHERE recording_id = 'recording:missing'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("failed ingestion left canonical state behind: count=%d", count)
	}
}
