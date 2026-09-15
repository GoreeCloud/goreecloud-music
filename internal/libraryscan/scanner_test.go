package libraryscan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestScanDiscoversSupportedAudioWithoutFollowingSymlinks(t *testing.T) {
	root := t.TempDir()
	mustWrite := func(relative, contents string) string {
		t.Helper()
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", relative, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", relative, err)
		}
		return path
	}

	mustWrite("zeta.MP3", "mp3")
	mustWrite("albums/alpha.flac", "flac")
	mustWrite("albums/beta.m4a", "m4a")
	mustWrite("notes.txt", "not audio")
	target := mustWrite("outside/linked.ogg", "ogg")
	if err := os.Symlink(target, filepath.Join(root, "linked.ogg")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	observations, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	got := make([]string, len(observations))
	for i, observation := range observations {
		got[i] = observation.RelativePath
		if observation.SizeBytes <= 0 {
			t.Fatalf("observation %q has size %d", observation.RelativePath, observation.SizeBytes)
		}
		if observation.ModifiedNS == 0 {
			t.Fatalf("observation %q has zero modified time", observation.RelativePath)
		}
	}
	want := []string{"albums/alpha.flac", "albums/beta.m4a", "outside/linked.ogg", "zeta.MP3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
}

func TestScanRejectsRelativeAndSymlinkRoots(t *testing.T) {
	if _, err := Scan("relative/library"); err == nil {
		t.Fatal("expected relative root to fail")
	}

	root := t.TempDir()
	linkedRoot := filepath.Join(t.TempDir(), "linked-root")
	if err := os.Symlink(root, linkedRoot); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := Scan(linkedRoot); err == nil {
		t.Fatal("expected symlink root to fail")
	}
}

func TestScanDoesNotModifySourceFiles(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "track.opus")
	before := []byte("source-media")
	if err := os.WriteFile(path, before, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Scan(root); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("source bytes changed: got %q want %q", after, before)
	}
}
