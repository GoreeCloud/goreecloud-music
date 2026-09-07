package playlists

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

const (
	PlaylistSchemaVersion   = 1
	MaxPlaylistPayloadBytes = 64 * 1024
)

var ErrInvalidPlaylistPayload = errors.New("invalid playlist payload")

type persistedPlaylist struct {
	SchemaVersion int      `json:"schemaVersion"`
	UserID        string   `json:"userId"`
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	TrackIDs      []string `json:"trackIds"`
}

// Encode serializes one validated playlist into the canonical versioned JSON
// shape used by future persistence adapters. It does not select a database or
// synchronization provider.
func Encode(playlist Playlist) ([]byte, error) {
	snapshot, err := playlist.Snapshot()
	if err != nil {
		return nil, ErrInvalidPlaylistPayload
	}
	payload, err := json.Marshal(persistedPlaylist{
		SchemaVersion: PlaylistSchemaVersion,
		UserID:        snapshot.UserID,
		ID:            snapshot.ID,
		Name:          snapshot.Name,
		TrackIDs:      append([]string{}, snapshot.TrackIDs...),
	})
	if err != nil || len(payload) == 0 || len(payload) > MaxPlaylistPayloadBytes {
		return nil, ErrInvalidPlaylistPayload
	}
	return payload, nil
}

// Decode accepts only the exact canonical schema produced by Encode. Unknown
// fields, trailing JSON, malformed state, and semantically equivalent but
// noncanonical JSON formatting fail closed.
func Decode(payload []byte) (Playlist, error) {
	if len(payload) == 0 || len(payload) > MaxPlaylistPayloadBytes {
		return Playlist{}, ErrInvalidPlaylistPayload
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var stored persistedPlaylist
	if err := decoder.Decode(&stored); err != nil {
		return Playlist{}, ErrInvalidPlaylistPayload
	}
	if err := requirePlaylistEOF(decoder); err != nil {
		return Playlist{}, ErrInvalidPlaylistPayload
	}
	if stored.SchemaVersion != PlaylistSchemaVersion {
		return Playlist{}, ErrInvalidPlaylistPayload
	}
	playlist, err := Restore(Snapshot{
		UserID:   stored.UserID,
		ID:       stored.ID,
		Name:     stored.Name,
		TrackIDs: stored.TrackIDs,
	})
	if err != nil {
		return Playlist{}, ErrInvalidPlaylistPayload
	}
	canonical, err := Encode(playlist)
	if err != nil || !bytes.Equal(canonical, payload) {
		return Playlist{}, ErrInvalidPlaylistPayload
	}
	return playlist, nil
}

func requirePlaylistEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrInvalidPlaylistPayload
	}
	return nil
}
