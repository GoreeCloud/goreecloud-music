package playlists

import (
	"errors"
	"strings"
)

const MaxPlaylistNameLength = 200

var ErrInvalidPlaylistState = errors.New("invalid playlist state")

type Snapshot struct {
	UserID   string
	ID       string
	Name     string
	TrackIDs []string
}

// Playlist is an ordered, user-scoped, persistence-neutral track collection.
// Duplicate track IDs are intentionally allowed because playlist order is a
// presentation choice rather than set membership.
type Playlist struct {
	userID   string
	id       string
	name     string
	trackIDs []string
}

func New(userID, id, name string) (Playlist, error) {
	if !validID(userID) || !validID(id) || !validName(name) {
		return Playlist{}, ErrInvalidPlaylistState
	}
	return Playlist{userID: userID, id: id, name: name, trackIDs: make([]string, 0)}, nil
}

func Restore(snapshot Snapshot) (Playlist, error) {
	playlist, err := New(snapshot.UserID, snapshot.ID, snapshot.Name)
	if err != nil {
		return Playlist{}, err
	}
	for _, trackID := range snapshot.TrackIDs {
		if !validID(trackID) {
			return Playlist{}, ErrInvalidPlaylistState
		}
		playlist.trackIDs = append(playlist.trackIDs, trackID)
	}
	return playlist, nil
}

func (p Playlist) UserID() string { return p.userID }
func (p Playlist) ID() string     { return p.id }
func (p Playlist) Name() string   { return p.name }
func (p Playlist) Len() int       { return len(p.trackIDs) }

func (p Playlist) TrackIDs() []string {
	if !validPlaylist(p) {
		return nil
	}
	return append([]string(nil), p.trackIDs...)
}

func (p *Playlist) Rename(name string) error {
	if p == nil || !validPlaylist(*p) || !validName(name) {
		return ErrInvalidPlaylistState
	}
	p.name = name
	return nil
}

func (p *Playlist) Append(trackID string) error {
	if p == nil || !validPlaylist(*p) || !validID(trackID) {
		return ErrInvalidPlaylistState
	}
	p.trackIDs = append(p.trackIDs, trackID)
	return nil
}

func (p *Playlist) Insert(index int, trackID string) error {
	if p == nil || !validPlaylist(*p) || !validID(trackID) || index < 0 || index > len(p.trackIDs) {
		return ErrInvalidPlaylistState
	}
	p.trackIDs = append(p.trackIDs, "")
	copy(p.trackIDs[index+1:], p.trackIDs[index:])
	p.trackIDs[index] = trackID
	return nil
}

func (p *Playlist) RemoveAt(index int) (string, error) {
	if p == nil || !validPlaylist(*p) || index < 0 || index >= len(p.trackIDs) {
		return "", ErrInvalidPlaylistState
	}
	removed := p.trackIDs[index]
	copy(p.trackIDs[index:], p.trackIDs[index+1:])
	p.trackIDs = p.trackIDs[:len(p.trackIDs)-1]
	return removed, nil
}

func (p *Playlist) Move(from, to int) error {
	if p == nil || !validPlaylist(*p) || from < 0 || from >= len(p.trackIDs) || to < 0 || to >= len(p.trackIDs) {
		return ErrInvalidPlaylistState
	}
	if from == to {
		return nil
	}
	trackID := p.trackIDs[from]
	if from < to {
		copy(p.trackIDs[from:to], p.trackIDs[from+1:to+1])
	} else {
		copy(p.trackIDs[to+1:from+1], p.trackIDs[to:from])
	}
	p.trackIDs[to] = trackID
	return nil
}

func (p Playlist) Snapshot() (Snapshot, error) {
	if !validPlaylist(p) {
		return Snapshot{}, ErrInvalidPlaylistState
	}
	return Snapshot{
		UserID:   p.userID,
		ID:       p.id,
		Name:     p.name,
		TrackIDs: append([]string(nil), p.trackIDs...),
	}, nil
}

func validPlaylist(p Playlist) bool {
	if !validID(p.userID) || !validID(p.id) || !validName(p.name) || p.trackIDs == nil {
		return false
	}
	for _, trackID := range p.trackIDs {
		if !validID(trackID) {
			return false
		}
	}
	return true
}

func validID(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}

func validName(value string) bool {
	return value != "" && len(value) <= MaxPlaylistNameLength && strings.TrimSpace(value) == value
}
