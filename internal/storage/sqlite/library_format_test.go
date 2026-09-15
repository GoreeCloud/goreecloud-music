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

func TestProbeLibraryFileFormatPersistsVerifiedMP3Facts(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "Artist", "Album", "Track.mp3")
	bytes := append(testID3v23(map[string]string{
		"TIT2": "Track",
		"TPE1": "Artist",
		"TALB": "Album",
		"TPE2": "Album Artist",
	}), testIngestMP3Frames()...)
	mustWriteLibraryFile(t, track, bytes)

	backend, owner, reader, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	ingest, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:format"), domain.ReleaseID("release:format"), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	got, err := backend.ProbeLibraryFileFormat(ctx, owner, library, fileID)
	if err != nil {
		t.Fatalf("ProbeLibraryFileFormat() error = %v", err)
	}
	if got.PlayableAssetID != ingest.PlayableAssetID || got.Codec != "mp3" || got.Container != "mpeg-audio" {
		t.Fatalf("format = %#v", got)
	}
	again, err := backend.ProbeLibraryFileFormat(ctx, owner, library, fileID)
	if err != nil || again != got {
		t.Fatalf("idempotent probe = %#v, err = %v", again, err)
	}

	var codec, container string
	var hash sql.NullString
	var duration sql.NullInt64
	if err := backend.db.QueryRowContext(ctx, `SELECT codec, container, media_sha256, duration_ms FROM playable_assets WHERE playable_asset_id = ?`, string(got.PlayableAssetID)).Scan(&codec, &container, &hash, &duration); err != nil {
		t.Fatal(err)
	}
	if codec != "mp3" || container != "mpeg-audio" || hash.Valid || duration.Valid {
		t.Fatalf("stored format = codec=%q container=%q hash=%#v duration=%#v", codec, container, hash, duration)
	}

	if _, err := backend.ProbeLibraryFileFormat(ctx, reader, library, fileID); err == nil {
		t.Fatal("expected read-only profile format probing to fail")
	}
}

func TestProbeLibraryFileFormatFailurePreservesExistingState(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	bytes := append(testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}), testIngestMP3Frames()...)
	mustWriteLibraryFile(t, track, bytes)
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	ingest, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:stale-format"), domain.ReleaseID("release:stale-format"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ProbeLibraryFileFormat(ctx, owner, library, fileID); err != nil {
		t.Fatal(err)
	}

	mustWriteLibraryFile(t, track, append(bytes, 0))
	if _, err := backend.ProbeLibraryFileFormat(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected probing changed source to fail")
	}
	var codec, container string
	if err := backend.db.QueryRowContext(ctx, `SELECT codec, container FROM playable_assets WHERE playable_asset_id = ?`, string(ingest.PlayableAssetID)).Scan(&codec, &container); err != nil {
		t.Fatal(err)
	}
	if codec != "mp3" || container != "mpeg-audio" {
		t.Fatalf("failed probe changed prior format state: %q / %q", codec, container)
	}
}

func TestProbeLibraryFileFormatRejectsTagOnlyMP3WithoutMutation(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mustWriteLibraryFile(t, track, testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}))
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	ingest, err := backend.IngestLibraryFile(ctx, owner, library, fileID, domain.RecordingID("recording:tag-only"), domain.ReleaseID("release:tag-only"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ProbeLibraryFileFormat(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected tag-only MP3 probing to fail")
	}
	var codec, container sql.NullString
	if err := backend.db.QueryRowContext(ctx, `SELECT codec, container FROM playable_assets WHERE playable_asset_id = ?`, string(ingest.PlayableAssetID)).Scan(&codec, &container); err != nil {
		t.Fatal(err)
	}
	if codec.Valid || container.Valid {
		t.Fatalf("failed probe persisted format state: %#v / %#v", codec, container)
	}
}

func TestProbeLibraryFileFormatRequiresMaterializedAsset(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	mustWriteLibraryFile(t, track, append(testID3v23(map[string]string{"TIT2": "Track", "TALB": "Album"}), testIngestMP3Frames()...))
	backend, owner, _, library, fileID := prepareIngestFixture(t, ctx, root)
	defer backend.Close()
	if _, err := backend.ProbeLibraryFileFormat(ctx, owner, library, fileID); err == nil {
		t.Fatal("expected probing before materialization to fail")
	}
	if err := os.Remove(track); err != nil {
		t.Fatal(err)
	}
}

func testIngestMP3Frames() []byte {
	const frameLength = 417
	frame := make([]byte, frameLength)
	copy(frame, []byte{0xff, 0xfb, 0x90, 0x64})
	return append(append([]byte{}, frame...), frame...)
}
