package artwork

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSidecarPrefersCover(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "folder.png"), []byte("png"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Cover.JPG"), []byte("jpg"), 0o600); err != nil {
		t.Fatal(err)
	}

	source, ok, err := DetectSidecar(context.Background(), audio)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected artwork")
	}
	if source.Type != SourceSidecar {
		t.Fatalf("type = %q", source.Type)
	}
	if filepath.Base(source.Path) != "Cover.JPG" {
		t.Fatalf("path = %q", source.Path)
	}
	if source.MIMEType != "image/jpeg" {
		t.Fatalf("mime = %q", source.MIMEType)
	}
}

func TestDetectSidecarMissing(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, ok, err := DetectSidecar(context.Background(), audio)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("unexpected artwork")
	}
}

func TestParseEmbeddedProbe(t *testing.T) {
	data := []byte(`{"streams":[{"codec_name":"mjpeg","disposition":{"attached_pic":1}}]}`)
	mimeType, found, err := ParseEmbeddedProbe(data)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected embedded artwork")
	}
	if mimeType != "image/jpeg" {
		t.Fatalf("mime = %q", mimeType)
	}
}

func TestParseEmbeddedProbeWithoutAttachedPicture(t *testing.T) {
	data := []byte(`{"streams":[{"codec_name":"png","disposition":{"attached_pic":0}}]}`)
	_, found, err := ParseEmbeddedProbe(data)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("unexpected embedded artwork")
	}
}
