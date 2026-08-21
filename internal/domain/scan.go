package domain

import (
	"errors"
	"strings"
)

type ScannedTrackFile struct {
	LibraryID   string
	Path        string
	SizeBytes   int64
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Genre       string
	ReleaseYear int
	TrackNumber int
	DiscNumber  int
	DurationMS  int64
	Codec       string
	Bitrate     int
	BitDepth    int
	SampleRate  int
	Channels    int
}

func (s ScannedTrackFile) Validate() error {
	if strings.TrimSpace(s.LibraryID) == "" {
		return errors.New("library id is required")
	}
	if strings.TrimSpace(s.Path) == "" {
		return errors.New("track file path is required")
	}
	if strings.TrimSpace(s.Title) == "" {
		return errors.New("track title is required")
	}
	if s.SizeBytes < 0 || s.DurationMS < 0 || s.Bitrate < 0 || s.BitDepth < 0 || s.SampleRate < 0 || s.Channels < 0 {
		return errors.New("technical metadata cannot be negative")
	}
	return nil
}
