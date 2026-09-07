package playlists

import (
	"bytes"
	"errors"
	"math"
)

type Revision uint64

var (
	ErrStaleRevision     = errors.New("stale playlist revision")
	ErrRevisionExhausted = errors.New("playlist revision exhausted")
	ErrInvalidRecord     = errors.New("invalid playlist record")
)

// Record is a storage-neutral optimistic-concurrency wrapper around one
// canonical playlist payload.
type Record struct {
	revision Revision
	payload  []byte
}

func NewRecord(playlist Playlist) (Record, error) {
	payload, err := Encode(playlist)
	if err != nil {
		return Record{}, ErrInvalidRecord
	}
	return Record{revision: 0, payload: append([]byte(nil), payload...)}, nil
}

func RestoreRecord(revision Revision, payload []byte) (Record, error) {
	if len(payload) == 0 || len(payload) > MaxPlaylistPayloadBytes {
		return Record{}, ErrInvalidRecord
	}
	if _, err := Decode(payload); err != nil {
		return Record{}, ErrInvalidRecord
	}
	return Record{revision: revision, payload: append([]byte(nil), payload...)}, nil
}

func (r Record) Revision() Revision { return r.revision }

func (r Record) Payload() []byte {
	return append([]byte(nil), r.payload...)
}

func (r Record) Restore() (Playlist, error) {
	playlist, err := Decode(r.payload)
	if err != nil {
		return Playlist{}, ErrInvalidRecord
	}
	return playlist, nil
}

// Update advances only the mutable playlist name/order while preserving the
// immutable user and playlist identities.
func (r Record) Update(expected Revision, playlist Playlist) (Record, error) {
	if expected != r.revision {
		return r, ErrStaleRevision
	}
	if r.revision == Revision(math.MaxUint64) {
		return r, ErrRevisionExhausted
	}
	current, err := r.Restore()
	if err != nil {
		return r, err
	}
	if !validPlaylist(playlist) || current.UserID() != playlist.UserID() || current.ID() != playlist.ID() {
		return r, ErrInvalidRecord
	}
	payload, err := Encode(playlist)
	if err != nil {
		return r, ErrInvalidRecord
	}
	return Record{
		revision: r.revision + 1,
		payload:  append([]byte(nil), payload...),
	}, nil
}

func sameRecord(left, right Record) bool {
	return left.Revision() == right.Revision() && bytes.Equal(left.Payload(), right.Payload())
}
