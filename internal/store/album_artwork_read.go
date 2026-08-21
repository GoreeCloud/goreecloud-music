package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GoreeCloud/goreecloud-music/internal/artwork"
)

func (s *PostgresStore) AlbumArtwork(ctx context.Context, libraryID, albumID string) (artwork.Source, error) {
	const query = `
		SELECT aa.source_type, aa.source_path, aa.mime_type
		FROM album_artwork aa
		JOIN albums a ON a.id = aa.album_id
		WHERE a.id::text = $1 AND a.library_id::text = $2`

	var source artwork.Source
	var sourceType string
	if err := s.db.QueryRowContext(ctx, query, albumID, libraryID).Scan(&sourceType, &source.Path, &source.MIMEType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return artwork.Source{}, ErrNotFound
		}
		return artwork.Source{}, fmt.Errorf("lookup album artwork: %w", err)
	}
	source.Type = artwork.SourceType(sourceType)
	if source.Type != artwork.SourceSidecar && source.Type != artwork.SourceEmbedded {
		return artwork.Source{}, fmt.Errorf("unsupported stored artwork source type %q", sourceType)
	}
	return source, nil
}
