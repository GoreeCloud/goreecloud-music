package playlists

import (
	"context"
	"errors"
)

var (
	ErrInvalidRepository       = errors.New("invalid playlist repository")
	ErrRecordNotFound          = errors.New("playlist record not found")
	ErrInvalidMutation         = errors.New("invalid playlist mutation")
	ErrInvalidRepositoryResult = errors.New("invalid playlist repository result")
)

// Repository is the storage-neutral persistence boundary for playlists.
// Implementations choose the actual durable store and must honor Record
// optimistic-revision semantics, including revision-checked deletion.
type Repository interface {
	Load(ctx context.Context, userID, playlistID string) (Record, bool, error)
	Create(ctx context.Context, playlist Playlist) (Record, error)
	Save(ctx context.Context, expected Revision, playlist Playlist) (Record, error)
	Delete(ctx context.Context, userID, playlistID string, expected Revision) (bool, error)
}

// RecordLister is the optional persisted catalog boundary used by services that
// can enumerate a user's independently versioned playlist records. Repositories
// that do not provide durable listing can still satisfy Repository without
// pretending that an in-memory catalog is authoritative after restart.
type RecordLister interface {
	List(ctx context.Context, userID string) ([]Record, error)
}

func InitializeStored(
	ctx context.Context,
	repository Repository,
	playlist Playlist,
) (Record, error) {
	if repository == nil || !validPlaylist(playlist) {
		return Record{}, ErrInvalidRepository
	}
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	expected, err := NewRecord(playlist)
	if err != nil {
		return Record{}, err
	}
	created, err := repository.Create(ctx, playlist)
	if err != nil {
		return Record{}, err
	}
	if !sameRecord(created, expected) {
		return Record{}, ErrInvalidRepositoryResult
	}
	return created, nil
}

func MutateStored(
	ctx context.Context,
	repository Repository,
	userID, playlistID string,
	mutate func(Playlist) (Playlist, error),
) (Record, error) {
	if repository == nil || mutate == nil || !validID(userID) || !validID(playlistID) {
		return Record{}, ErrInvalidMutation
	}
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	record, found, err := repository.Load(ctx, userID, playlistID)
	if err != nil {
		return Record{}, err
	}
	if !found {
		return Record{}, ErrRecordNotFound
	}
	current, err := record.Restore()
	if err != nil || current.UserID() != userID || current.ID() != playlistID {
		return Record{}, ErrInvalidRepositoryResult
	}

	nextPlaylist, err := mutate(current)
	if err != nil {
		return record, err
	}
	expected, err := record.Update(record.Revision(), nextPlaylist)
	if err != nil {
		return record, err
	}
	if err := ctx.Err(); err != nil {
		return record, err
	}

	saved, err := repository.Save(ctx, record.Revision(), nextPlaylist)
	if err != nil {
		return record, err
	}
	if !sameRecord(saved, expected) {
		return record, ErrInvalidRepositoryResult
	}
	return saved, nil
}

// DeleteStored deletes exactly the record that was loaded. A concurrent writer
// that advances the optimistic revision wins; the stale delete must fail rather
// than erasing the newer playlist state.
func DeleteStored(
	ctx context.Context,
	repository Repository,
	userID, playlistID string,
) (Record, error) {
	if repository == nil || !validID(userID) || !validID(playlistID) {
		return Record{}, ErrInvalidMutation
	}
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	record, found, err := repository.Load(ctx, userID, playlistID)
	if err != nil {
		return Record{}, err
	}
	if !found {
		return Record{}, ErrRecordNotFound
	}
	playlist, err := record.Restore()
	if err != nil || playlist.UserID() != userID || playlist.ID() != playlistID {
		return Record{}, ErrInvalidRepositoryResult
	}
	if err := ctx.Err(); err != nil {
		return record, err
	}
	deleted, err := repository.Delete(ctx, userID, playlistID, record.Revision())
	if err != nil {
		return record, err
	}
	if !deleted {
		return record, ErrInvalidRepositoryResult
	}
	return record, nil
}
