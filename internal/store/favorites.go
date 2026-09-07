package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

var ErrInvalidTrackFavoriteStore = errors.New("invalid track favorite store")

func (s *PostgresStore) TrackForLibrary(ctx context.Context, libraryID, trackID string) (domain.Track, error) {
	if s == nil || s.db == nil || !validFavoriteID(libraryID) || !validFavoriteID(trackID) {
		return domain.Track{}, ErrInvalidTrackFavoriteStore
	}
	const query = `
		SELECT id::text, library_id::text, COALESCE(album_id::text, ''), title,
		       track_number, disc_number, duration_ms, created_at
		FROM tracks
		WHERE library_id::text = $1 AND id::text = $2`
	var track domain.Track
	if err := s.db.QueryRowContext(ctx, query, libraryID, trackID).Scan(
		&track.ID,
		&track.LibraryID,
		&track.AlbumID,
		&track.Title,
		&track.TrackNumber,
		&track.DiscNumber,
		&track.DurationMS,
		&track.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Track{}, ErrNotFound
		}
		return domain.Track{}, err
	}
	return track, nil
}

func (s *PostgresStore) TrackFavoritesForUser(ctx context.Context, userID string) ([]domain.Track, error) {
	if s == nil || s.db == nil || !validFavoriteID(userID) {
		return nil, ErrInvalidTrackFavoriteStore
	}
	const query = `
		SELECT t.id::text, t.library_id::text, COALESCE(t.album_id::text, ''), t.title,
		       t.track_number, t.disc_number, t.duration_ms, t.created_at
		FROM track_favorites f
		JOIN tracks t ON t.id = f.track_id
		WHERE f.user_id::text = $1
		ORDER BY f.created_at DESC, t.id`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]domain.Track, 0)
	for rows.Next() {
		var track domain.Track
		if err := rows.Scan(
			&track.ID,
			&track.LibraryID,
			&track.AlbumID,
			&track.Title,
			&track.TrackNumber,
			&track.DiscNumber,
			&track.DurationMS,
			&track.CreatedAt,
		); err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tracks, nil
}

func (s *PostgresStore) SetTrackFavorite(ctx context.Context, userID, trackID string, favorite bool) error {
	if s == nil || s.db == nil || !validFavoriteID(userID) || !validFavoriteID(trackID) {
		return ErrInvalidTrackFavoriteStore
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if favorite {
		const query = `
			INSERT INTO track_favorites (user_id, track_id)
			VALUES ($1::uuid, $2::uuid)
			ON CONFLICT (user_id, track_id) DO NOTHING`
		_, err := s.db.ExecContext(ctx, query, userID, trackID)
		return err
	}
	const query = `DELETE FROM track_favorites WHERE user_id::text = $1 AND track_id::text = $2`
	_, err := s.db.ExecContext(ctx, query, userID, trackID)
	return err
}

func validFavoriteID(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}
