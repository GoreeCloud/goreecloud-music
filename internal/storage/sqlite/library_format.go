package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/media"
)

// PlayableAssetFormat is the bounded codec/container result persisted for an
// already-materialized GoreeCloud Server playable asset.
type PlayableAssetFormat struct {
	PlayableAssetID domain.PlayableAssetID
	Codec           string
	Container       string
}

// ProbeLibraryFileFormat validates the current source bytes for an already
// materialized library file and persists only codec/container facts. It does
// not create canonical identity, infer equivalence, hash media, or claim
// duration/bitrate/sample-rate/channel facts. Library edit permission is
// re-evaluated inside the transaction.
func (b *Backend) ProbeLibraryFileFormat(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, fileID string) (PlayableAssetFormat, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return PlayableAssetFormat{}, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return PlayableAssetFormat{}, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" || len(fileID) > 256 {
		return PlayableAssetFormat{}, fmt.Errorf("library file id must contain 1 to 256 characters")
	}
	sourceItemID, err := localSourceItemID(validatedLibrary, fileID)
	if err != nil {
		return PlayableAssetFormat{}, err
	}
	assetID, err := localPlayableAssetID(validatedLibrary, fileID)
	if err != nil {
		return PlayableAssetFormat{}, err
	}

	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return PlayableAssetFormat{}, err
	}
	defer func() { _ = tx.Rollback() }()

	allowed, err := txHasLibraryEditPermission(ctx, tx, validatedProfile, validatedLibrary)
	if err != nil {
		return PlayableAssetFormat{}, err
	}
	if !allowed {
		return PlayableAssetFormat{}, fmt.Errorf("library edit permission required")
	}

	var rootPath, relativePath string
	var sizeBytes, modifiedNS int64
	var missingSince sql.NullString
	var sourceRecording, assetRecording, mediaPath string
	var mediaSize sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT l.root_path, f.relative_path, f.size_bytes, f.modified_ns, f.missing_since,
		s.recording_id, a.recording_id, a.media_path, a.media_size_bytes
		FROM library_files f
		JOIN libraries l ON l.library_id = f.library_id
		JOIN source_items s ON s.source_item_id = ? AND s.library_id = f.library_id AND s.source_kind = ? AND s.external_id = f.file_id
		JOIN playable_assets a ON a.playable_asset_id = ? AND a.library_id = f.library_id AND a.source_item_id = s.source_item_id
		WHERE f.file_id = ? AND f.library_id = ?`,
		string(sourceItemID), string(domain.SourceGoreeCloudServer), string(assetID), fileID, string(validatedLibrary)).Scan(
		&rootPath, &relativePath, &sizeBytes, &modifiedNS, &missingSince,
		&sourceRecording, &assetRecording, &mediaPath, &mediaSize)
	if errors.Is(err, sql.ErrNoRows) {
		return PlayableAssetFormat{}, fmt.Errorf("materialized playable asset is required before format probing")
	}
	if err != nil {
		return PlayableAssetFormat{}, fmt.Errorf("load playable asset for format probing: %w", err)
	}
	if missingSince.Valid {
		return PlayableAssetFormat{}, fmt.Errorf("library file is currently missing")
	}
	if sourceRecording == "" || sourceRecording != assetRecording || mediaPath != relativePath || !mediaSize.Valid || mediaSize.Int64 != sizeBytes {
		return PlayableAssetFormat{}, fmt.Errorf("playable asset binding no longer matches scanner/source state")
	}

	source, err := openObservedLibraryFile(rootPath, relativePath, sizeBytes, modifiedNS)
	if err != nil {
		return PlayableAssetFormat{}, err
	}
	defer source.Close()

	format, err := media.ProbeFormat(source, filepath.Ext(relativePath))
	if err != nil {
		return PlayableAssetFormat{}, fmt.Errorf("probe media format for %q: %w", relativePath, err)
	}
	if err := verifyObservedLibraryFileSnapshot(source, sizeBytes, modifiedNS); err != nil {
		return PlayableAssetFormat{}, err
	}

	result, err := tx.ExecContext(ctx, `UPDATE playable_assets SET codec = ?, container = ?
		WHERE playable_asset_id = ? AND library_id = ? AND recording_id = ? AND source_item_id = ? AND media_path = ? AND media_size_bytes = ?`,
		format.Codec, format.Container, string(assetID), string(validatedLibrary), assetRecording, string(sourceItemID), mediaPath, sizeBytes)
	if err != nil {
		return PlayableAssetFormat{}, fmt.Errorf("persist playable asset format: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return PlayableAssetFormat{}, fmt.Errorf("confirm playable asset format update: %w", err)
	}
	if rows != 1 {
		return PlayableAssetFormat{}, fmt.Errorf("playable asset changed during format probing")
	}
	if err := tx.Commit(); err != nil {
		return PlayableAssetFormat{}, fmt.Errorf("commit playable asset format: %w", err)
	}
	return PlayableAssetFormat{PlayableAssetID: assetID, Codec: format.Codec, Container: format.Container}, nil
}
