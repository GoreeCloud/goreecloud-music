package sqlitestore

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func TestExtractLibraryFileMetadataPersistsAndTracksStaleness(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "Artist", "Track.mp3")
	mustWriteLibraryFile(t, track, testID3v23(map[string]string{
		"TIT2": "First Title",
		"TPE1": "Artist",
		"TALB": "Album",
		"TRCK": "1/10",
		"TYER": "2026",
	}))

	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	owner := domain.ProfileID("profile:owner")
	reader := domain.ProfileID("profile:reader")
	outsider := domain.ProfileID("profile:outsider")
	library := domain.LibraryID("library:metadata")
	for _, profile := range []domain.ProfileID{owner, reader, outsider} {
		if err := backend.CreateProfile(ctx, profile, string(profile)); err != nil {
			t.Fatal(err)
		}
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Metadata", root); err != nil {
		t.Fatal(err)
	}
	if err := backend.SetLibraryPermission(ctx, library, reader, LibraryPermissionRead); err != nil {
		t.Fatal(err)
	}

	firstScan := time.Date(2026, 9, 15, 21, 10, 0, 0, time.UTC)
	if _, err := backend.ScanLibrary(ctx, owner, library, firstScan); err != nil {
		t.Fatal(err)
	}
	files, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(files) != 1 {
		t.Fatalf("files = %#v, err = %v", files, err)
	}
	fileID := files[0].ID

	firstExtract := firstScan.Add(time.Minute)
	stored, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, fileID, firstExtract)
	if err != nil {
		t.Fatalf("ExtractLibraryFileMetadata() error = %v", err)
	}
	if stored.Title != "First Title" || stored.Artist != "Artist" || stored.Album != "Album" || stored.TagFormat != "id3v2.3" || !stored.Current {
		t.Fatalf("stored metadata = %#v", stored)
	}

	visible, found, err := backend.LibraryFileMetadataForProfile(ctx, reader, library, fileID)
	if err != nil || !found {
		t.Fatalf("reader metadata = %#v, found = %v, err = %v", visible, found, err)
	}
	if visible.Title != "First Title" || !visible.Current {
		t.Fatalf("reader metadata = %#v", visible)
	}
	if _, err := backend.ExtractLibraryFileMetadata(ctx, reader, library, fileID, time.Now()); err == nil {
		t.Fatal("expected read-only profile extraction to fail")
	}
	if _, _, err := backend.LibraryFileMetadataForProfile(ctx, outsider, library, fileID); err == nil {
		t.Fatal("expected unauthorized metadata read to fail")
	}

	mustWriteLibraryFile(t, track, testID3v23(map[string]string{
		"TIT2": "Second Title With Different Size",
		"TPE1": "Artist",
		"TALB": "Album",
	}))
	if _, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, fileID, time.Now()); err == nil {
		t.Fatal("expected extraction against stale scanner observation to fail")
	}

	secondScan := firstScan.Add(2 * time.Minute)
	result, err := backend.ScanLibrary(ctx, owner, library, secondScan)
	if err != nil {
		t.Fatal(err)
	}
	if result.Updated != 1 {
		t.Fatalf("scan result = %#v, want one updated file", result)
	}
	stale, found, err := backend.LibraryFileMetadataForProfile(ctx, owner, library, fileID)
	if err != nil || !found {
		t.Fatalf("stale metadata = %#v, found = %v, err = %v", stale, found, err)
	}
	if stale.Current {
		t.Fatalf("metadata should be stale after scanner observation changed: %#v", stale)
	}

	updated, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, fileID, secondScan.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Second Title With Different Size" || !updated.Current {
		t.Fatalf("updated metadata = %#v", updated)
	}

	if err := os.Remove(track); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ScanLibrary(ctx, owner, library, secondScan.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	missing, found, err := backend.LibraryFileMetadataForProfile(ctx, owner, library, fileID)
	if err != nil || !found {
		t.Fatalf("missing metadata = %#v, found = %v, err = %v", missing, found, err)
	}
	if missing.Current {
		t.Fatalf("metadata must not be current for missing file: %#v", missing)
	}
	if _, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, fileID, time.Now()); err == nil {
		t.Fatal("expected metadata extraction for missing file to fail")
	}
}

func TestExtractLibraryFileMetadataRejectsSymlinkSubstitution(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	track := filepath.Join(root, "track.mp3")
	contents := testID3v23(map[string]string{"TIT2": "Original"})
	mustWriteLibraryFile(t, track, contents)

	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	owner := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:symlink")
	if err := backend.CreateProfile(ctx, owner, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, owner, "Symlink", root); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ScanLibrary(ctx, owner, library, time.Now()); err != nil {
		t.Fatal(err)
	}
	files, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(files) != 1 {
		t.Fatalf("files = %#v, err = %v", files, err)
	}

	target := filepath.Join(t.TempDir(), "target.mp3")
	mustWriteLibraryFile(t, target, contents)
	if err := os.Remove(track); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, track); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, files[0].ID, time.Now()); err == nil {
		t.Fatal("expected symlink substitution to fail")
	}
}

func TestExtractLibraryFileMetadataRejectsUnsupportedContainer(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	mustWriteLibraryFile(t, filepath.Join(root, "track.m4a"), []byte("m4a-placeholder"))
	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	owner := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:m4a")
	if err := backend.CreateProfile(ctx, owner, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, owner, "M4A", root); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.ScanLibrary(ctx, owner, library, time.Now()); err != nil {
		t.Fatal(err)
	}
	files, err := backend.LibraryFilesForProfile(ctx, owner, library, false)
	if err != nil || len(files) != 1 {
		t.Fatalf("files = %#v, err = %v", files, err)
	}
	if _, err := backend.ExtractLibraryFileMetadata(ctx, owner, library, files[0].ID, time.Now()); err == nil {
		t.Fatal("expected unsupported M4A metadata extraction to fail")
	}
}

func testID3v23(values map[string]string) []byte {
	order := []string{"TIT2", "TPE1", "TALB", "TPE2", "TCON", "TRCK", "TPOS", "TYER"}
	var payload bytes.Buffer
	for _, id := range order {
		value := values[id]
		if value == "" {
			continue
		}
		data := append([]byte{3}, []byte(value)...)
		payload.WriteString(id)
		_ = binary.Write(&payload, binary.BigEndian, uint32(len(data)))
		payload.Write([]byte{0, 0})
		payload.Write(data)
	}
	header := []byte{'I', 'D', '3', 3, 0, 0}
	size := payload.Len()
	header = append(header,
		byte((size>>21)&0x7f),
		byte((size>>14)&0x7f),
		byte((size>>7)&0x7f),
		byte(size&0x7f),
	)
	return append(header, payload.Bytes()...)
}
