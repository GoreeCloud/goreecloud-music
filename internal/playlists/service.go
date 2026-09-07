package playlists

import (
	"context"
	"errors"
)

var ErrInvalidService = errors.New("invalid playlist service")

type Service struct {
	repository Repository
}

func NewService(repository Repository) (Service, error) {
	if repository == nil {
		return Service{}, ErrInvalidService
	}
	return Service{repository: repository}, nil
}

func (s Service) Create(ctx context.Context, userID, playlistID, name string) (Record, error) {
	if s.repository == nil {
		return Record{}, ErrInvalidService
	}
	playlist, err := New(userID, playlistID, name)
	if err != nil {
		return Record{}, err
	}
	return InitializeStored(ctx, s.repository, playlist)
}

func (s Service) Load(ctx context.Context, userID, playlistID string) (Record, Playlist, error) {
	if s.repository == nil || !validID(userID) || !validID(playlistID) {
		return Record{}, Playlist{}, ErrInvalidService
	}
	if err := ctx.Err(); err != nil {
		return Record{}, Playlist{}, err
	}
	record, found, err := s.repository.Load(ctx, userID, playlistID)
	if err != nil {
		return Record{}, Playlist{}, err
	}
	if !found {
		return Record{}, Playlist{}, ErrRecordNotFound
	}
	playlist, err := record.Restore()
	if err != nil || playlist.UserID() != userID || playlist.ID() != playlistID {
		return Record{}, Playlist{}, ErrInvalidRepositoryResult
	}
	return record, playlist, nil
}

func (s Service) Rename(ctx context.Context, userID, playlistID, name string) (Record, error) {
	return s.mutate(ctx, userID, playlistID, func(playlist *Playlist) error {
		return playlist.Rename(name)
	})
}

func (s Service) Append(ctx context.Context, userID, playlistID, trackID string) (Record, error) {
	return s.mutate(ctx, userID, playlistID, func(playlist *Playlist) error {
		return playlist.Append(trackID)
	})
}

func (s Service) Insert(ctx context.Context, userID, playlistID string, index int, trackID string) (Record, error) {
	return s.mutate(ctx, userID, playlistID, func(playlist *Playlist) error {
		return playlist.Insert(index, trackID)
	})
}

func (s Service) RemoveAt(ctx context.Context, userID, playlistID string, index int) (Record, string, error) {
	var removed string
	record, err := s.mutate(ctx, userID, playlistID, func(playlist *Playlist) error {
		var err error
		removed, err = playlist.RemoveAt(index)
		return err
	})
	if err != nil {
		return record, "", err
	}
	return record, removed, nil
}

func (s Service) Move(ctx context.Context, userID, playlistID string, from, to int) (Record, error) {
	return s.mutate(ctx, userID, playlistID, func(playlist *Playlist) error {
		return playlist.Move(from, to)
	})
}

func (s Service) mutate(
	ctx context.Context,
	userID, playlistID string,
	mutate func(*Playlist) error,
) (Record, error) {
	if s.repository == nil || mutate == nil {
		return Record{}, ErrInvalidService
	}
	return MutateStored(ctx, s.repository, userID, playlistID, func(current Playlist) (Playlist, error) {
		if err := mutate(&current); err != nil {
			return current, err
		}
		return current, nil
	})
}
