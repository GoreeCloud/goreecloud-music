package ingest

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/metadata"
)

type fakeRepository struct {
	items       []domain.ScannedTrackFile
	staleCalled bool
	removed     int64
}

func (f *fakeRepository) UpsertScannedTrackFile(_ context.Context, item domain.ScannedTrackFile, _ time.Time) error {
	f.items = append(f.items, item)
	return nil
}

func (f *fakeRepository) DeleteStaleTrackFiles(_ context.Context, _ string, _ time.Time) (int64, error) {
	f.staleCalled = true
	return f.removed, nil
}

func TestScanLibraryIndexesSupportedFilesAndReconciles(t *testing.T) {
	root := t.TempDir()
	musicPath := filepath.Join(root, "signal.flac")
	if err := os.WriteFile(musicPath, []byte("audio"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "ignore.txt"), []byte("text"), 0o600); err != nil {
		t.Fatal(err)
	}

	repository := &fakeRepository{removed: 2}
	service := New(repository, func(_ context.Context, path string) (metadata.TrackMetadata, error) {
		if path != musicPath {
			t.Fatalf("unexpected probe path %q", path)
		}
		return metadata.TrackMetadata{
			Title:       "Signal",
			Artist:      "Example Artist",
			Album:       "Example Album",
			AlbumArtist: "Example Artist",
			Genre:       "Electronic",
			Date:        "2026-08-21",
			Track:       "4/10",
			Disc:        "1/1",
			Codec:       "flac",
			Bitrate:     921600,
			BitDepth:    24,
			SampleRate:  96000,
			Channels:    2,
			DurationMS:  245125,
		}, nil
	})
	service.now = func() time.Time { return time.Date(2026, 8, 21, 16, 0, 0, 0, time.UTC) }

	summary, err := service.ScanLibrary(context.Background(), domain.Library{
		ID:         "11111111-1111-4111-8111-111111111111",
		Name:       "Shared Music",
		RootPath:   root,
		Visibility: domain.LibraryShared,
	})
	if err != nil {
		t.Fatalf("ScanLibrary returned error: %v", err)
	}
	if summary.Discovered != 1 || summary.Indexed != 1 || summary.Failed != 0 || summary.Removed != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if !repository.staleCalled || len(repository.items) != 1 {
		t.Fatalf("unexpected repository state: %+v", repository)
	}
	item := repository.items[0]
	if item.Title != "Signal" || item.TrackNumber != 4 || item.DiscNumber != 1 || item.ReleaseYear != 2026 || item.BitDepth != 24 {
		t.Fatalf("unexpected scanned item: %+v", item)
	}
}

func TestScanLibrarySkipsDestructiveCleanupAfterProbeFailure(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "broken.mp3"), []byte("audio"), 0o600); err != nil {
		t.Fatal(err)
	}

	repository := &fakeRepository{}
	service := New(repository, func(context.Context, string) (metadata.TrackMetadata, error) {
		return metadata.TrackMetadata{}, errors.New("probe failed")
	})

	summary, err := service.ScanLibrary(context.Background(), domain.Library{
		ID:         "11111111-1111-4111-8111-111111111111",
		Name:       "Shared Music",
		RootPath:   root,
		Visibility: domain.LibraryShared,
	})
	if err == nil {
		t.Fatal("expected scan error")
	}
	if summary.Failed != 1 || repository.staleCalled {
		t.Fatalf("cleanup should be skipped after scan failure: summary=%+v repo=%+v", summary, repository)
	}
}
