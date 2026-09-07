package playlists

import (
	"errors"
	"sort"
)

var ErrInvalidPlaylistCatalog = errors.New("invalid playlist catalog")

type CatalogEntry struct {
	ID         string
	Name       string
	TrackCount int
}

type CatalogSnapshot struct {
	UserID    string
	Playlists []Snapshot
}

// Catalog is a user-scoped, persistence-neutral collection of independently
// ordered playlists. Playlist IDs are unique inside one user's catalog.
type Catalog struct {
	userID    string
	playlists map[string]Playlist
}

func NewCatalog(userID string) (Catalog, error) {
	if !validID(userID) {
		return Catalog{}, ErrInvalidPlaylistCatalog
	}
	return Catalog{userID: userID, playlists: make(map[string]Playlist)}, nil
}

func RestoreCatalog(snapshot CatalogSnapshot) (Catalog, error) {
	catalog, err := NewCatalog(snapshot.UserID)
	if err != nil {
		return Catalog{}, err
	}
	for _, playlistSnapshot := range snapshot.Playlists {
		playlist, err := Restore(playlistSnapshot)
		if err != nil || playlist.UserID() != snapshot.UserID {
			return Catalog{}, ErrInvalidPlaylistCatalog
		}
		if _, exists := catalog.playlists[playlist.ID()]; exists {
			return Catalog{}, ErrInvalidPlaylistCatalog
		}
		catalog.playlists[playlist.ID()] = playlist
	}
	return catalog, nil
}

func (c Catalog) UserID() string { return c.userID }
func (c Catalog) Count() int     { return len(c.playlists) }

func (c *Catalog) Add(playlist Playlist) error {
	if c == nil || !validCatalog(*c) || !validPlaylist(playlist) || playlist.UserID() != c.userID {
		return ErrInvalidPlaylistCatalog
	}
	if _, exists := c.playlists[playlist.ID()]; exists {
		return ErrInvalidPlaylistCatalog
	}
	clone, err := clonePlaylist(playlist)
	if err != nil {
		return ErrInvalidPlaylistCatalog
	}
	c.playlists[clone.ID()] = clone
	return nil
}

func (c *Catalog) Remove(id string) (bool, error) {
	if c == nil || !validCatalog(*c) || !validID(id) {
		return false, ErrInvalidPlaylistCatalog
	}
	if _, exists := c.playlists[id]; !exists {
		return false, nil
	}
	delete(c.playlists, id)
	return true, nil
}

func (c Catalog) Get(id string) (Playlist, bool) {
	if !validCatalog(c) || !validID(id) {
		return Playlist{}, false
	}
	playlist, exists := c.playlists[id]
	if !exists {
		return Playlist{}, false
	}
	clone, err := clonePlaylist(playlist)
	if err != nil {
		return Playlist{}, false
	}
	return clone, true
}

// Entries returns deterministic playlist-ID ordering and metadata-only copies
// suitable for library/list presentation without exposing mutable track slices.
func (c Catalog) Entries() []CatalogEntry {
	if !validCatalog(c) {
		return nil
	}
	ids := make([]string, 0, len(c.playlists))
	for id := range c.playlists {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	entries := make([]CatalogEntry, 0, len(ids))
	for _, id := range ids {
		playlist := c.playlists[id]
		entries = append(entries, CatalogEntry{ID: id, Name: playlist.Name(), TrackCount: playlist.Len()})
	}
	return entries
}

func (c Catalog) Snapshot() (CatalogSnapshot, error) {
	if !validCatalog(c) {
		return CatalogSnapshot{}, ErrInvalidPlaylistCatalog
	}
	ids := make([]string, 0, len(c.playlists))
	for id := range c.playlists {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	playlists := make([]Snapshot, 0, len(ids))
	for _, id := range ids {
		snapshot, err := c.playlists[id].Snapshot()
		if err != nil {
			return CatalogSnapshot{}, ErrInvalidPlaylistCatalog
		}
		playlists = append(playlists, snapshot)
	}
	return CatalogSnapshot{UserID: c.userID, Playlists: playlists}, nil
}

func validCatalog(c Catalog) bool {
	if !validID(c.userID) || c.playlists == nil {
		return false
	}
	for id, playlist := range c.playlists {
		if id != playlist.ID() || playlist.UserID() != c.userID || !validPlaylist(playlist) {
			return false
		}
	}
	return true
}

func clonePlaylist(playlist Playlist) (Playlist, error) {
	snapshot, err := playlist.Snapshot()
	if err != nil {
		return Playlist{}, err
	}
	return Restore(snapshot)
}
