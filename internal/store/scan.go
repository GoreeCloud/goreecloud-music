package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/id"
)

func (s *PostgresStore) LibraryByID(ctx context.Context, libraryID string) (domain.Library, error) {
	const query = `
		SELECT id::text, name, root_path, visibility, created_at
		FROM libraries
		WHERE id::text = $1`

	var library domain.Library
	if err := s.db.QueryRowContext(ctx, query, libraryID).Scan(
		&library.ID,
		&library.Name,
		&library.RootPath,
		&library.Visibility,
		&library.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Library{}, ErrNotFound
		}
		return domain.Library{}, err
	}
	return library, nil
}

func (s *PostgresStore) UpsertScannedTrackFile(ctx context.Context, item domain.ScannedTrackFile, seenAt time.Time) error {
	if err := item.Validate(); err != nil {
		return err
	}
	if seenAt.IsZero() {
		seenAt = time.Now().UTC()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin scanned track upsert: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	albumArtistName := strings.TrimSpace(item.AlbumArtist)
	if albumArtistName == "" {
		albumArtistName = strings.TrimSpace(item.Artist)
	}

	artistID, err := ensureArtist(ctx, tx, albumArtistName)
	if err != nil {
		return err
	}
	albumID, err := ensureAlbum(ctx, tx, item.LibraryID, strings.TrimSpace(item.Album), artistID, item.ReleaseYear)
	if err != nil {
		return err
	}

	var fileID, trackID, existingLibraryID string
	err = tx.QueryRowContext(ctx, `
		SELECT tf.id::text, tf.track_id::text, t.library_id::text
		FROM track_files tf
		JOIN tracks t ON t.id = tf.track_id
		WHERE tf.path = $1`, item.Path).Scan(&fileID, &trackID, &existingLibraryID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("lookup scanned track file: %w", err)
	}
	if err == nil && existingLibraryID != item.LibraryID {
		return fmt.Errorf("track file path already belongs to another library")
	}

	albumValue := nullableUUID(albumID)
	if errors.Is(err, sql.ErrNoRows) {
		trackID, err = id.NewUUID()
		if err != nil {
			return fmt.Errorf("create track id: %w", err)
		}
		fileID, err = id.NewUUID()
		if err != nil {
			return fmt.Errorf("create track file id: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO tracks (
				id, library_id, album_id, title, track_number, disc_number,
				duration_ms, artist_name, genre, release_year
			) VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9, $10)`,
			trackID,
			item.LibraryID,
			albumValue,
			item.Title,
			item.TrackNumber,
			item.DiscNumber,
			item.DurationMS,
			item.Artist,
			item.Genre,
			nullableInt(item.ReleaseYear),
		)
		if err != nil {
			return fmt.Errorf("insert scanned track: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO track_files (
				id, track_id, path, codec, bitrate, bit_depth, sample_rate,
				channels, size_bytes, last_seen_at, updated_at
			) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, now())`,
			fileID,
			trackID,
			item.Path,
			item.Codec,
			item.Bitrate,
			item.BitDepth,
			item.SampleRate,
			item.Channels,
			item.SizeBytes,
			seenAt,
		)
		if err != nil {
			return fmt.Errorf("insert scanned track file: %w", err)
		}
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE tracks
			SET album_id = $2::uuid,
				title = $3,
				track_number = $4,
				disc_number = $5,
				duration_ms = $6,
				artist_name = $7,
				genre = $8,
				release_year = $9
			WHERE id::text = $1`,
			trackID,
			albumValue,
			item.Title,
			item.TrackNumber,
			item.DiscNumber,
			item.DurationMS,
			item.Artist,
			item.Genre,
			nullableInt(item.ReleaseYear),
		)
		if err != nil {
			return fmt.Errorf("update scanned track: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE track_files
			SET codec = $2,
				bitrate = $3,
				bit_depth = $4,
				sample_rate = $5,
				channels = $6,
				size_bytes = $7,
				last_seen_at = $8,
				updated_at = now()
			WHERE id::text = $1`,
			fileID,
			item.Codec,
			item.Bitrate,
			item.BitDepth,
			item.SampleRate,
			item.Channels,
			item.SizeBytes,
			seenAt,
		)
		if err != nil {
			return fmt.Errorf("update scanned track file: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit scanned track upsert: %w", err)
	}
	return nil
}

func (s *PostgresStore) DeleteStaleTrackFiles(ctx context.Context, libraryID string, before time.Time) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin stale track cleanup: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		DELETE FROM track_files tf
		USING tracks t
		WHERE tf.track_id = t.id
			AND t.library_id::text = $1
			AND tf.last_seen_at < $2`, libraryID, before)
	if err != nil {
		return 0, fmt.Errorf("delete stale track files: %w", err)
	}
	removed, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count stale track files: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM tracks t
		WHERE t.library_id::text = $1
			AND NOT EXISTS (
				SELECT 1 FROM track_files tf WHERE tf.track_id = t.id
			)`, libraryID); err != nil {
		return 0, fmt.Errorf("delete orphan tracks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit stale track cleanup: %w", err)
	}
	return removed, nil
}

func ensureArtist(ctx context.Context, tx *sql.Tx, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil
	}

	var artistID string
	err := tx.QueryRowContext(ctx,
		`SELECT id::text FROM artists WHERE lower(name) = lower($1) ORDER BY id LIMIT 1`,
		name,
	).Scan(&artistID)
	if err == nil {
		return artistID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("lookup album artist: %w", err)
	}

	artistID, err = id.NewUUID()
	if err != nil {
		return "", fmt.Errorf("create artist id: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO artists (id, name, sort_name)
		VALUES ($1::uuid, $2, $2)
		ON CONFLICT ((lower(name))) DO NOTHING`, artistID, name); err != nil {
		return "", fmt.Errorf("insert album artist: %w", err)
	}

	if err := tx.QueryRowContext(ctx,
		`SELECT id::text FROM artists WHERE lower(name) = lower($1) ORDER BY id LIMIT 1`,
		name,
	).Scan(&artistID); err != nil {
		return "", fmt.Errorf("resolve album artist: %w", err)
	}
	return artistID, nil
}

func ensureAlbum(ctx context.Context, tx *sql.Tx, libraryID, title, artistID string, releaseYear int) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", nil
	}

	var albumID string
	err := tx.QueryRowContext(ctx, `
		SELECT id::text
		FROM albums
		WHERE library_id::text = $1
			AND title = $2
			AND (($3 = '' AND album_artist_id IS NULL) OR album_artist_id::text = $3)
		ORDER BY id
		LIMIT 1`, libraryID, title, artistID).Scan(&albumID)
	if err == nil {
		if releaseYear > 0 {
			_, _ = tx.ExecContext(ctx, `UPDATE albums SET release_year = COALESCE(release_year, $2) WHERE id::text = $1`, albumID, releaseYear)
		}
		return albumID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("lookup album: %w", err)
	}

	albumID, err = id.NewUUID()
	if err != nil {
		return "", fmt.Errorf("create album id: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO albums (id, library_id, title, album_artist_id, release_year)
		VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5)`,
		albumID,
		libraryID,
		title,
		nullableUUID(artistID),
		nullableInt(releaseYear),
	); err != nil {
		return "", fmt.Errorf("insert album: %w", err)
	}
	return albumID, nil
}

func nullableUUID(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableInt(value int) any {
	if value <= 0 {
		return nil
	}
	return value
}
