package preferences

import (
	"errors"
	"sort"
	"strings"
)

var ErrInvalidFavoritesState = errors.New("invalid favorites state")

// FavoriteTracks is a user-scoped, persistence-neutral set of exact track IDs.
// User and track identifiers remain opaque; production GoreeCloud Identity and
// durable storage integration are separate adapter responsibilities.
type FavoriteTracks struct {
	userID   string
	trackIDs map[string]struct{}
}

type FavoriteTracksSnapshot struct {
	UserID   string
	TrackIDs []string
}

func NewFavoriteTracks(userID string) (FavoriteTracks, error) {
	if !validID(userID) {
		return FavoriteTracks{}, ErrInvalidFavoritesState
	}
	return FavoriteTracks{userID: userID, trackIDs: make(map[string]struct{})}, nil
}

// RestoreFavoriteTracks accepts only exact canonical snapshot state. Duplicate,
// blank, or whitespace-altered identifiers fail closed rather than being
// silently normalized or deduplicated.
func RestoreFavoriteTracks(snapshot FavoriteTracksSnapshot) (FavoriteTracks, error) {
	favorites, err := NewFavoriteTracks(snapshot.UserID)
	if err != nil {
		return FavoriteTracks{}, err
	}
	for _, trackID := range snapshot.TrackIDs {
		if !validID(trackID) {
			return FavoriteTracks{}, ErrInvalidFavoritesState
		}
		if _, exists := favorites.trackIDs[trackID]; exists {
			return FavoriteTracks{}, ErrInvalidFavoritesState
		}
		favorites.trackIDs[trackID] = struct{}{}
	}
	return favorites, nil
}

func (f FavoriteTracks) UserID() string { return f.userID }
func (f FavoriteTracks) Count() int     { return len(f.trackIDs) }

func (f FavoriteTracks) Contains(trackID string) bool {
	if !validID(trackID) || !validFavoriteTracks(f) {
		return false
	}
	_, ok := f.trackIDs[trackID]
	return ok
}

func (f *FavoriteTracks) Add(trackID string) (bool, error) {
	if f == nil || !validFavoriteTracks(*f) || !validID(trackID) {
		return false, ErrInvalidFavoritesState
	}
	if _, ok := f.trackIDs[trackID]; ok {
		return false, nil
	}
	f.trackIDs[trackID] = struct{}{}
	return true, nil
}

func (f *FavoriteTracks) Remove(trackID string) (bool, error) {
	if f == nil || !validFavoriteTracks(*f) || !validID(trackID) {
		return false, ErrInvalidFavoritesState
	}
	if _, ok := f.trackIDs[trackID]; !ok {
		return false, nil
	}
	delete(f.trackIDs, trackID)
	return true, nil
}

// Snapshot returns a deterministic, defensively copied persistence shape.
func (f FavoriteTracks) Snapshot() (FavoriteTracksSnapshot, error) {
	if !validFavoriteTracks(f) {
		return FavoriteTracksSnapshot{}, ErrInvalidFavoritesState
	}
	ids := make([]string, 0, len(f.trackIDs))
	for id := range f.trackIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return FavoriteTracksSnapshot{UserID: f.userID, TrackIDs: ids}, nil
}

func validFavoriteTracks(f FavoriteTracks) bool {
	if !validID(f.userID) || f.trackIDs == nil {
		return false
	}
	for id := range f.trackIDs {
		if !validID(id) {
			return false
		}
	}
	return true
}

func validID(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}
