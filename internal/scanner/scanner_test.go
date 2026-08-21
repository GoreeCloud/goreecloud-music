package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFindsSupportedAudioOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	album := filepath.Join(root, "Artist", "Album")
	if err := os.MkdirAll(album, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"01 - Track.FLAC": "audio",
		"02 - Track.mp3":  "audio2",
		"cover.jpg":       "image",
		"notes.txt":       "notes",
	} {
		if err := os.WriteFile(filepath.Join(album, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	files, err := Discover(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 audio files, got %d", len(files))
	}
	if files[0].Extension != ".flac" || files[1].Extension != ".mp3" {
		t.Fatalf("unexpected extensions: %#v", files)
	}
}

func TestDiscoverRejectsFileRoot(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "track.flac")
	if err := os.WriteFile(path, []byte("audio"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(path); err != ErrInvalidRoot {
		t.Fatalf("expected ErrInvalidRoot, got %v", err)
	}
}
