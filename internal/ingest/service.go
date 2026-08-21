package ingest

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/artwork"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/metadata"
	"github.com/GoreeCloud/goreecloud-music/internal/scanner"
)

type Repository interface {
	UpsertScannedTrackFile(ctx context.Context, item domain.ScannedTrackFile, seenAt time.Time) error
	DeleteStaleTrackFiles(ctx context.Context, libraryID string, before time.Time) (int64, error)
}

type ArtworkRepository interface {
	UpsertAlbumArtworkForTrackPath(ctx context.Context, libraryID, trackPath string, source artwork.Source) error
}

type ProbeFunc func(context.Context, string) (metadata.TrackMetadata, error)
type ArtworkDetector func(context.Context, string) (artwork.Source, bool, error)

type Summary struct {
	Discovered int   `json:"discovered"`
	Indexed    int   `json:"indexed"`
	Artwork    int   `json:"artwork"`
	Removed    int64 `json:"removed"`
	Failed     int   `json:"failed"`
}

type Service struct {
	repository    Repository
	probe         ProbeFunc
	detectArtwork ArtworkDetector
	now           func() time.Time
}

func New(repository Repository, probe ProbeFunc) *Service {
	return &Service{
		repository:    repository,
		probe:         probe,
		detectArtwork: artwork.Detect,
		now:           time.Now,
	}
}

func (s *Service) ScanLibrary(ctx context.Context, library domain.Library) (Summary, error) {
	if s == nil || s.repository == nil || s.probe == nil {
		return Summary{}, errors.New("ingestion service is not configured")
	}
	if err := library.Validate(); err != nil {
		return Summary{}, err
	}

	files, err := scanner.Discover(library.RootPath)
	if err != nil {
		return Summary{}, fmt.Errorf("discover library files: %w", err)
	}

	scanStarted := s.now().UTC()
	summary := Summary{Discovered: len(files)}
	var failures []error
	artworkRepository, persistArtwork := s.repository.(ArtworkRepository)

	for _, file := range files {
		trackMetadata, err := s.probe(ctx, file.Path)
		if err != nil {
			summary.Failed++
			failures = append(failures, fmt.Errorf("probe %s: %w", file.Path, err))
			continue
		}

		item := scannedTrackFile(library.ID, file, trackMetadata)
		if err := s.repository.UpsertScannedTrackFile(ctx, item, scanStarted); err != nil {
			summary.Failed++
			failures = append(failures, fmt.Errorf("persist %s: %w", file.Path, err))
			continue
		}
		summary.Indexed++

		if persistArtwork && s.detectArtwork != nil {
			source, found, err := s.detectArtwork(ctx, file.Path)
			if err != nil {
				summary.Failed++
				failures = append(failures, fmt.Errorf("discover artwork %s: %w", file.Path, err))
				continue
			}
			if found {
				if err := artworkRepository.UpsertAlbumArtworkForTrackPath(ctx, library.ID, file.Path, source); err != nil {
					summary.Failed++
					failures = append(failures, fmt.Errorf("persist artwork %s: %w", file.Path, err))
					continue
				}
				summary.Artwork++
			}
		}
	}

	if len(failures) > 0 {
		return summary, errors.Join(failures...)
	}

	removed, err := s.repository.DeleteStaleTrackFiles(ctx, library.ID, scanStarted)
	if err != nil {
		return summary, fmt.Errorf("remove stale track files: %w", err)
	}
	summary.Removed = removed
	return summary, nil
}

func scannedTrackFile(libraryID string, file scanner.File, trackMetadata metadata.TrackMetadata) domain.ScannedTrackFile {
	title := strings.TrimSpace(trackMetadata.Title)
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))
	}

	return domain.ScannedTrackFile{
		LibraryID:   libraryID,
		Path:        file.Path,
		SizeBytes:   file.SizeBytes,
		Title:       title,
		Artist:      strings.TrimSpace(trackMetadata.Artist),
		Album:       strings.TrimSpace(trackMetadata.Album),
		AlbumArtist: strings.TrimSpace(trackMetadata.AlbumArtist),
		Genre:       strings.TrimSpace(trackMetadata.Genre),
		ReleaseYear: leadingNumber(trackMetadata.Date),
		TrackNumber: leadingNumber(trackMetadata.Track),
		DiscNumber:  leadingNumber(trackMetadata.Disc),
		DurationMS:  trackMetadata.DurationMS,
		Codec:       strings.TrimSpace(trackMetadata.Codec),
		Bitrate:     trackMetadata.Bitrate,
		BitDepth:    trackMetadata.BitDepth,
		SampleRate:  trackMetadata.SampleRate,
		Channels:    trackMetadata.Channels,
	}
}

func leadingNumber(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	end := 0
	for end < len(value) && value[end] >= '0' && value[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0
	}
	parsed, _ := strconv.Atoi(value[:end])
	return parsed
}
