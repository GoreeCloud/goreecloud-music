package store

import (
	"context"

	"github.com/GoreeCloud/goreecloud-music/internal/playlists"
)

// Delete removes exactly one user-scoped playlist record only when its current
// optimistic revision still equals expected. A newer concurrent save therefore
// wins over a stale delete.
func (s *PostgresStore) Delete(
	ctx context.Context,
	userID, playlistID string,
	expected playlists.Revision,
) (bool, error) {
	if s == nil || s.db == nil || userID == "" || playlistID == "" {
		return false, ErrInvalidPostgresPlaylistStore
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	const query = `
		DELETE FROM playlist_records
		WHERE user_id = $1 AND playlist_id = $2 AND revision = $3`
	result, err := s.db.ExecContext(ctx, query, userID, playlistID, formatPlaylistRevision(expected))
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows == 1 {
		return true, nil
	}
	if rows != 0 {
		return false, playlists.ErrInvalidRepositoryResult
	}
	current, found, err := s.Load(ctx, userID, playlistID)
	if err != nil {
		return false, err
	}
	if !found {
		return false, playlists.ErrRecordNotFound
	}
	if current.Revision() != expected {
		return false, playlists.ErrStaleRevision
	}
	return false, playlists.ErrInvalidRepositoryResult
}
