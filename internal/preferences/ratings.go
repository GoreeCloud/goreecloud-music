package preferences

import (
	"errors"
	"sort"
)

const (
	MinTrackRating = 1
	MaxTrackRating = 5
)

var ErrInvalidTrackRatingsState = errors.New("invalid track ratings state")

type TrackRating struct {
	TrackID string
	Stars   int
}

type TrackRatingsSnapshot struct {
	UserID  string
	Ratings []TrackRating
}

// TrackRatings is a user-scoped, persistence-neutral map of exact track IDs to
// one-through-five star values. Zero is reserved for the caller-facing
// unrated/cleared state and is never persisted as a rating.
type TrackRatings struct {
	userID  string
	ratings map[string]int
}

func NewTrackRatings(userID string) (TrackRatings, error) {
	if !validID(userID) {
		return TrackRatings{}, ErrInvalidTrackRatingsState
	}
	return TrackRatings{userID: userID, ratings: make(map[string]int)}, nil
}

// RestoreTrackRatings accepts only exact canonical snapshot state. Duplicate
// track IDs, whitespace-altered identifiers, and out-of-range star values fail
// closed rather than being silently normalized.
func RestoreTrackRatings(snapshot TrackRatingsSnapshot) (TrackRatings, error) {
	ratings, err := NewTrackRatings(snapshot.UserID)
	if err != nil {
		return TrackRatings{}, err
	}
	for _, rating := range snapshot.Ratings {
		if !validID(rating.TrackID) || !validStars(rating.Stars) {
			return TrackRatings{}, ErrInvalidTrackRatingsState
		}
		if _, exists := ratings.ratings[rating.TrackID]; exists {
			return TrackRatings{}, ErrInvalidTrackRatingsState
		}
		ratings.ratings[rating.TrackID] = rating.Stars
	}
	return ratings, nil
}

func (r TrackRatings) UserID() string { return r.userID }
func (r TrackRatings) Count() int     { return len(r.ratings) }

func (r TrackRatings) Rating(trackID string) (int, bool) {
	if !validTrackRatings(r) || !validID(trackID) {
		return 0, false
	}
	stars, ok := r.ratings[trackID]
	return stars, ok
}

func (r *TrackRatings) Set(trackID string, stars int) (bool, error) {
	if r == nil || !validTrackRatings(*r) || !validID(trackID) || !validStars(stars) {
		return false, ErrInvalidTrackRatingsState
	}
	previous, exists := r.ratings[trackID]
	if exists && previous == stars {
		return false, nil
	}
	r.ratings[trackID] = stars
	return true, nil
}

func (r *TrackRatings) Clear(trackID string) (bool, error) {
	if r == nil || !validTrackRatings(*r) || !validID(trackID) {
		return false, ErrInvalidTrackRatingsState
	}
	if _, exists := r.ratings[trackID]; !exists {
		return false, nil
	}
	delete(r.ratings, trackID)
	return true, nil
}

// Snapshot returns deterministic track-ID ordering and a fresh rating slice so
// callers cannot mutate the live preference state through persistence values.
func (r TrackRatings) Snapshot() (TrackRatingsSnapshot, error) {
	if !validTrackRatings(r) {
		return TrackRatingsSnapshot{}, ErrInvalidTrackRatingsState
	}
	ids := make([]string, 0, len(r.ratings))
	for trackID := range r.ratings {
		ids = append(ids, trackID)
	}
	sort.Strings(ids)
	values := make([]TrackRating, 0, len(ids))
	for _, trackID := range ids {
		values = append(values, TrackRating{TrackID: trackID, Stars: r.ratings[trackID]})
	}
	return TrackRatingsSnapshot{UserID: r.userID, Ratings: values}, nil
}

func validTrackRatings(r TrackRatings) bool {
	if !validID(r.userID) || r.ratings == nil {
		return false
	}
	for trackID, stars := range r.ratings {
		if !validID(trackID) || !validStars(stars) {
			return false
		}
	}
	return true
}

func validStars(stars int) bool {
	return stars >= MinTrackRating && stars <= MaxTrackRating
}
