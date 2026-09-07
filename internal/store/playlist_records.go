package store

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"

	"github.com/GoreeCloud/goreecloud-music/internal/playlists"
)

var (
	ErrInvalidPostgresPlaylistStore = errors.New("invalid postgres playlist store")
	ErrPlaylistAlreadyExists        = errors.New("playlist already exists")
)

func (s *PostgresStore) Load(ctx context.Context, userID, playlistID string) (playlists.Record, bool, error) {
	if s == nil || s.db == nil || userID == "" || playlistID == "" {
		return playlists.Record{}, false, ErrInvalidPostgresPlaylistStore
	}
	if err := ctx.Err(); err != nil {
		return playlists.Record{}, false, err
	}
	const query = `
		SELECT revision::text, payload
		FROM playlist_records
		WHERE user_id = $1 AND playlist_id = $2`
	record, err := scanPlaylistRecord(s.db.QueryRowContext(ctx, query, userID, playlistID))
	if errors.Is(err, sql.ErrNoRows) {
		return playlists.Record{}, false, nil
	}
	if err != nil {
		return playlists.Record{}, false, err
	}
	if err := validatePlaylistRecordIdentity(record, userID, playlistID); err != nil {
		return playlists.Record{}, false, err
	}
	return record, true, nil
}

func (s *PostgresStore) Create(ctx context.Context, playlist playlists.Playlist) (playlists.Record, error) {
	if s == nil || s.db == nil {
		return playlists.Record{}, ErrInvalidPostgresPlaylistStore
	}
	if err := ctx.Err(); err != nil {
		return playlists.Record{}, err
	}
	expected, err := playlists.NewRecord(playlist)
	if err != nil {
		return playlists.Record{}, err
	}
	const query = `
		INSERT INTO playlist_records (user_id, playlist_id, revision, payload)
		VALUES ($1, $2, 0, $3)
		ON CONFLICT (user_id, playlist_id) DO NOTHING
		RETURNING revision::text, payload`
	created, err := scanPlaylistRecord(s.db.QueryRowContext(ctx, query, playlist.UserID(), playlist.ID(), expected.Payload()))
	if errors.Is(err, sql.ErrNoRows) {
		return playlists.Record{}, ErrPlaylistAlreadyExists
	}
	if err != nil {
		return playlists.Record{}, err
	}
	if !samePlaylistRecord(created, expected) {
		return playlists.Record{}, playlists.ErrInvalidRepositoryResult
	}
	return created, nil
}

func (s *PostgresStore) Save(
	ctx context.Context,
	expectedRevision playlists.Revision,
	playlist playlists.Playlist,
) (playlists.Record, error) {
	if s == nil || s.db == nil {
		return playlists.Record{}, ErrInvalidPostgresPlaylistStore
	}
	if err := ctx.Err(); err != nil {
		return playlists.Record{}, err
	}
	if expectedRevision == playlists.Revision(math.MaxUint64) {
		return playlists.Record{}, playlists.ErrRevisionExhausted
	}
	payload, err := playlists.Encode(playlist)
	if err != nil {
		return playlists.Record{}, err
	}
	nextRevision := expectedRevision + 1
	const query = `
		UPDATE playlist_records
		SET revision = $1, payload = $2, updated_at = NOW()
		WHERE user_id = $3 AND playlist_id = $4 AND revision = $5
		RETURNING revision::text, payload`
	updated, err := scanPlaylistRecord(s.db.QueryRowContext(
		ctx,
		query,
		formatPlaylistRevision(nextRevision),
		payload,
		playlist.UserID(),
		playlist.ID(),
		formatPlaylistRevision(expectedRevision),
	))
	if err == nil {
		expected, restoreErr := playlists.RestoreRecord(nextRevision, payload)
		if restoreErr != nil {
			return playlists.Record{}, restoreErr
		}
		if !samePlaylistRecord(updated, expected) {
			return playlists.Record{}, playlists.ErrInvalidRepositoryResult
		}
		return updated, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return playlists.Record{}, err
	}

	current, found, loadErr := s.Load(ctx, playlist.UserID(), playlist.ID())
	if loadErr != nil {
		return playlists.Record{}, loadErr
	}
	if !found {
		return playlists.Record{}, playlists.ErrRecordNotFound
	}
	return current, playlists.ErrStaleRevision
}

type playlistRecordScanner interface {
	Scan(dest ...any) error
}

func scanPlaylistRecord(row playlistRecordScanner) (playlists.Record, error) {
	var revisionText string
	var payload []byte
	if err := row.Scan(&revisionText, &payload); err != nil {
		return playlists.Record{}, err
	}
	revision, err := parsePlaylistRevision(revisionText)
	if err != nil {
		return playlists.Record{}, playlists.ErrInvalidRecord
	}
	record, err := playlists.RestoreRecord(revision, payload)
	if err != nil {
		return playlists.Record{}, playlists.ErrInvalidRecord
	}
	return record, nil
}

func validatePlaylistRecordIdentity(record playlists.Record, userID, playlistID string) error {
	playlist, err := record.Restore()
	if err != nil || playlist.UserID() != userID || playlist.ID() != playlistID {
		return playlists.ErrInvalidRepositoryResult
	}
	return nil
}

func samePlaylistRecord(left, right playlists.Record) bool {
	return left.Revision() == right.Revision() && bytes.Equal(left.Payload(), right.Payload())
}

func formatPlaylistRevision(revision playlists.Revision) string {
	return strconv.FormatUint(uint64(revision), 10)
}

func parsePlaylistRevision(value string) (playlists.Revision, error) {
	if value == "" || value[0] == '+' || (len(value) > 1 && value[0] == '0') {
		return 0, playlists.ErrInvalidRecord
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, playlists.ErrInvalidRecord
	}
	if formatPlaylistRevision(playlists.Revision(parsed)) != value {
		return 0, playlists.ErrInvalidRecord
	}
	return playlists.Revision(parsed), nil
}
