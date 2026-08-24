package domain

import (
	"context"
	"errors"
)

var ErrTrackNotStreamable = errors.New("track is not streamable")

type ByteRange struct {
	Start int64
	End   *int64
}

type StreamRequest struct {
	ActorID   string
	LibraryID string
	TrackID   string
	Range     *ByteRange
}

type StreamDescriptor struct {
	ContentType   string
	ContentLength int64
	AcceptRanges  bool
	ETag          string
	SourcePath    string
}

// StreamAuthority is the first-party Resonance Audio boundary for resolving an authorized
// track into a local source descriptor. Implementations must enforce library membership before
// returning a source path and must never expose arbitrary filesystem paths from request input.
type StreamAuthority interface {
	Resolve(ctx context.Context, request StreamRequest) (StreamDescriptor, error)
}

func (r StreamRequest) Validate() error {
	if r.ActorID == "" || r.LibraryID == "" || r.TrackID == "" {
		return ErrTrackNotStreamable
	}
	if r.Range != nil {
		if r.Range.Start < 0 {
			return ErrTrackNotStreamable
		}
		if r.Range.End != nil && *r.Range.End < r.Range.Start {
			return ErrTrackNotStreamable
		}
	}
	return nil
}
