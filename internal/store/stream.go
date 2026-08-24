package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime"
	"path/filepath"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func (s *PostgresStore) ResolveTrackStream(ctx context.Context, libraryID, trackID string) (domain.StreamDescriptor, error) {
	const query = `
		SELECT tf.path, tf.size_bytes
		FROM track_files tf
		JOIN tracks t ON t.id = tf.track_id
		WHERE t.library_id::text = $1
			AND t.id::text = $2
		ORDER BY tf.updated_at DESC, tf.id
		LIMIT 1`

	var path string
	var size int64
	if err := s.db.QueryRowContext(ctx, query, libraryID, trackID).Scan(&path, &size); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.StreamDescriptor{}, ErrNotFound
		}
		return domain.StreamDescriptor{}, fmt.Errorf("resolve track stream: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return domain.StreamDescriptor{
		ContentType:   contentType,
		ContentLength: size,
		AcceptRanges:  true,
		SourcePath:    path,
	}, nil
}
