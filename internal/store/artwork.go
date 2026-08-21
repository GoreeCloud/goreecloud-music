package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/artwork"
)

func (s *PostgresStore) UpsertAlbumArtworkForTrackPath(ctx context.Context, libraryID, trackPath string, source artwork.Source) error {
	libraryID = strings.TrimSpace(libraryID)
	trackPath = strings.TrimSpace(trackPath)
	if libraryID == "" || trackPath == "" {
		return errors.New("library id and track path are required")
	}
	if source.Type != artwork.SourceSidecar && source.Type != artwork.SourceEmbedded {
		return errors.New("unsupported artwork source type")
	}
	if strings.TrimSpace(source.Path) == "" {
		return errors.New("artwork source path is required")
	}

	var albumID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT t.album_id::text
		FROM track_files tf
		JOIN tracks t ON t.id = tf.track_id
		WHERE tf.path = $1 AND t.library_id::text = $2`, trackPath, libraryID).Scan(&albumID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("lookup artwork album: %w", err)
	}
	if !albumID.Valid || strings.TrimSpace(albumID.String) == "" {
		return nil
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO album_artwork (album_id, source_type, source_path, mime_type, updated_at)
		VALUES ($1::uuid, $2, $3, $4, now())
		ON CONFLICT (album_id) DO UPDATE SET
			source_type = EXCLUDED.source_type,
			source_path = EXCLUDED.source_path,
			mime_type = EXCLUDED.mime_type,
			updated_at = now()`,
		albumID.String, string(source.Type), source.Path, source.MIMEType)
	if err != nil {
		return fmt.Errorf("upsert album artwork: %w", err)
	}
	return nil
}
