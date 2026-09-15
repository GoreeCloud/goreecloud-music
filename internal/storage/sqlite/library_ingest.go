package sqlitestore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

// LibraryFileIngestResult identifies the canonical and source-local objects
// associated with one explicitly authorized library-file ingestion. Recording
// and Release identity are caller-supplied; source/asset identity is derived
// only from the durable scanner file identity.
type LibraryFileIngestResult struct {
	RecordingID          domain.RecordingID
	ReleaseID            domain.ReleaseID
	SourceItemID         domain.SourceItemID
	PlayableAssetID      domain.PlayableAssetID
	CreatedRecording     bool
	CreatedRelease       bool
	CreatedSourceItem    bool
	CreatedPlayableAsset bool
}

// IngestLibraryFile materializes one current scanner/metadata observation into
// canonical Music storage using explicit Recording and Release identities. It
// deliberately does not infer that similarly tagged files are equivalent.
// Library edit permission is re-evaluated inside the transaction.
func (b *Backend) IngestLibraryFile(
	ctx context.Context,
	profileID domain.ProfileID,
	libraryID domain.LibraryID,
	fileID string,
	recordingID domain.RecordingID,
	releaseID domain.ReleaseID,
	ingestedAt time.Time,
) (LibraryFileIngestResult, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	validatedRecording, err := domain.NewRecordingID(string(recordingID))
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	validatedRelease, err := domain.NewReleaseID(string(releaseID))
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" || len(fileID) > 256 {
		return LibraryFileIngestResult{}, fmt.Errorf("library file id must contain 1 to 256 characters")
	}
	if ingestedAt.IsZero() {
		ingestedAt = time.Now().UTC()
	} else {
		ingestedAt = ingestedAt.UTC()
	}

	sourceItemID, err := localSourceItemID(validatedLibrary, fileID)
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	assetID, err := localPlayableAssetID(validatedLibrary, fileID)
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	result := LibraryFileIngestResult{
		RecordingID:     validatedRecording,
		ReleaseID:       validatedRelease,
		SourceItemID:    sourceItemID,
		PlayableAssetID: assetID,
	}

	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	allowed, err := txHasLibraryEditPermission(ctx, tx, validatedProfile, validatedLibrary)
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	if !allowed {
		return LibraryFileIngestResult{}, fmt.Errorf("library edit permission required")
	}

	var rootPath, relativePath string
	var sizeBytes, modifiedNS, metadataSize, metadataModified int64
	var missingSince sql.NullString
	var title, artist, album, albumArtist string
	err = tx.QueryRowContext(ctx, `SELECT l.root_path, f.relative_path, f.size_bytes, f.modified_ns, f.missing_since,
		m.title, m.artist, m.album, m.album_artist, m.source_size_bytes, m.source_modified_ns
		FROM library_files f
		JOIN libraries l ON l.library_id = f.library_id
		JOIN library_file_metadata m ON m.file_id = f.file_id AND m.library_id = f.library_id
		WHERE f.file_id = ? AND f.library_id = ?`, fileID, string(validatedLibrary)).Scan(
		&rootPath, &relativePath, &sizeBytes, &modifiedNS, &missingSince,
		&title, &artist, &album, &albumArtist, &metadataSize, &metadataModified)
	if errors.Is(err, sql.ErrNoRows) {
		return LibraryFileIngestResult{}, fmt.Errorf("current extracted metadata is required before ingestion")
	}
	if err != nil {
		return LibraryFileIngestResult{}, fmt.Errorf("load library file metadata for ingestion: %w", err)
	}
	if missingSince.Valid || metadataSize != sizeBytes || metadataModified != modifiedNS {
		return LibraryFileIngestResult{}, fmt.Errorf("library file metadata is stale; rescan and re-extract before ingestion")
	}

	title = strings.TrimSpace(title)
	artist = strings.TrimSpace(artist)
	album = strings.TrimSpace(album)
	albumArtist = strings.TrimSpace(albumArtist)
	if title == "" {
		return LibraryFileIngestResult{}, fmt.Errorf("current metadata must include a title before canonical ingestion")
	}
	if album == "" {
		return LibraryFileIngestResult{}, fmt.Errorf("current metadata must include an album before release ingestion")
	}
	if albumArtist == "" {
		albumArtist = artist
	}

	// Reconfirm the actual filesystem object still matches the scanner snapshot.
	// The open remains read-only and uses the rooted/symlink-safe helper shared
	// with metadata extraction. Size/mtime is freshness evidence, not a content hash.
	source, err := openObservedLibraryFile(rootPath, relativePath, sizeBytes, modifiedNS)
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	defer source.Close()

	created, err := ensureReleaseForIngest(ctx, tx, validatedLibrary, validatedRelease, album, albumArtist)
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	result.CreatedRelease = created

	created, err = ensureRecordingForIngest(ctx, tx, validatedLibrary, validatedRecording, validatedRelease, title, artist, ingestedAt)
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	result.CreatedRecording = created

	created, err = ensureSourceItemForIngest(ctx, tx, validatedLibrary, sourceItemID, validatedRecording, fileID)
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	result.CreatedSourceItem = created

	created, err = ensurePlayableAssetForIngest(ctx, tx, validatedLibrary, assetID, validatedRecording, sourceItemID, relativePath, sizeBytes, ingestedAt)
	if err != nil {
		return LibraryFileIngestResult{}, err
	}
	result.CreatedPlayableAsset = created

	// Close the gap between the initial rooted open and the durable commit.
	// A source modified while this transaction is materializing application
	// state must fail closed rather than leave an asset bound to stale facts.
	if err := verifyObservedLibraryFileSnapshot(source, sizeBytes, modifiedNS); err != nil {
		return LibraryFileIngestResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return LibraryFileIngestResult{}, fmt.Errorf("commit library file ingestion: %w", err)
	}
	return result, nil
}

func txHasLibraryEditPermission(ctx context.Context, tx *sql.Tx, profileID domain.ProfileID, libraryID domain.LibraryID) (bool, error) {
	var permission string
	err := tx.QueryRowContext(ctx, `SELECT permission FROM library_memberships WHERE library_id = ? AND profile_id = ?`, string(libraryID), string(profileID)).Scan(&permission)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("load library permission for ingestion: %w", err)
	}
	return permission == string(LibraryPermissionEdit) || permission == string(LibraryPermissionOwner), nil
}

func ensureReleaseForIngest(ctx context.Context, tx *sql.Tx, libraryID domain.LibraryID, releaseID domain.ReleaseID, title, artist string) (bool, error) {
	var existingLibrary, existingTitle, existingArtist string
	err := tx.QueryRowContext(ctx, `SELECT library_id, title, artist FROM releases WHERE release_id = ?`, string(releaseID)).Scan(&existingLibrary, &existingTitle, &existingArtist)
	if err == nil {
		if existingLibrary != string(libraryID) || existingTitle != title || existingArtist != artist {
			return false, fmt.Errorf("release id %q conflicts with existing canonical release", releaseID)
		}
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("load release for ingestion: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO releases(release_id, library_id, title, artist, release_date, metadata_json) VALUES(?, ?, ?, ?, NULL, '{}')`, string(releaseID), string(libraryID), title, artist); err != nil {
		return false, fmt.Errorf("create release for ingestion: %w", err)
	}
	return true, nil
}

func ensureRecordingForIngest(ctx context.Context, tx *sql.Tx, libraryID domain.LibraryID, recordingID domain.RecordingID, releaseID domain.ReleaseID, title, artist string, addedAt time.Time) (bool, error) {
	var existingLibrary, existingTitle, existingArtist string
	var existingRelease sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT library_id, title, artist, release_id FROM recordings WHERE recording_id = ?`, string(recordingID)).Scan(&existingLibrary, &existingTitle, &existingArtist, &existingRelease)
	if err == nil {
		if existingLibrary != string(libraryID) || existingTitle != title || existingArtist != artist || !existingRelease.Valid || existingRelease.String != string(releaseID) {
			return false, fmt.Errorf("recording id %q conflicts with existing canonical recording", recordingID)
		}
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("load recording for ingestion: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO recordings(recording_id, library_id, title, artist, release_id, metadata_json, added_at) VALUES(?, ?, ?, ?, ?, '{}', ?)`, string(recordingID), string(libraryID), title, artist, string(releaseID), addedAt.Format(time.RFC3339Nano)); err != nil {
		return false, fmt.Errorf("create recording for ingestion: %w", err)
	}
	return true, nil
}

func ensureSourceItemForIngest(ctx context.Context, tx *sql.Tx, libraryID domain.LibraryID, sourceItemID domain.SourceItemID, recordingID domain.RecordingID, fileID string) (bool, error) {
	var existingID, existingRecording string
	err := tx.QueryRowContext(ctx, `SELECT source_item_id, recording_id FROM source_items WHERE library_id = ? AND source_kind = ? AND external_id = ? LIMIT 1`, string(libraryID), string(domain.SourceGoreeCloudServer), fileID).Scan(&existingID, &existingRecording)
	if err == nil && (existingID != string(sourceItemID) || existingRecording != string(recordingID)) {
		return false, fmt.Errorf("library file %q conflicts with existing GoreeCloud Server source item", fileID)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("load source item by library file: %w", err)
	}

	var existingLibrary, existingKind string
	var existingRecordingID, existingExternal sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT library_id, recording_id, source_kind, external_id FROM source_items WHERE source_item_id = ?`, string(sourceItemID)).Scan(&existingLibrary, &existingRecordingID, &existingKind, &existingExternal)
	if err == nil {
		if existingLibrary != string(libraryID) || !existingRecordingID.Valid || existingRecordingID.String != string(recordingID) || existingKind != string(domain.SourceGoreeCloudServer) || !existingExternal.Valid || existingExternal.String != fileID {
			return false, fmt.Errorf("source item id %q conflicts with existing source item", sourceItemID)
		}
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("load source item for ingestion: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO source_items(source_item_id, library_id, recording_id, source_kind, external_id, metadata_json) VALUES(?, ?, ?, ?, ?, '{}')`, string(sourceItemID), string(libraryID), string(recordingID), string(domain.SourceGoreeCloudServer), fileID); err != nil {
		return false, fmt.Errorf("create source item for ingestion: %w", err)
	}
	return true, nil
}

func ensurePlayableAssetForIngest(ctx context.Context, tx *sql.Tx, libraryID domain.LibraryID, assetID domain.PlayableAssetID, recordingID domain.RecordingID, sourceItemID domain.SourceItemID, relativePath string, sizeBytes int64, addedAt time.Time) (bool, error) {
	var existingID, existingRecording string
	err := tx.QueryRowContext(ctx, `SELECT playable_asset_id, recording_id FROM playable_assets WHERE library_id = ? AND media_path = ? LIMIT 1`, string(libraryID), relativePath).Scan(&existingID, &existingRecording)
	if err == nil && (existingID != string(assetID) || existingRecording != string(recordingID)) {
		return false, fmt.Errorf("media path %q conflicts with existing playable asset", relativePath)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("load playable asset by path: %w", err)
	}

	var existingLibrary, existingRecordingID, existingPath string
	var existingSource sql.NullString
	var existingSize sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT library_id, recording_id, source_item_id, media_path, media_size_bytes FROM playable_assets WHERE playable_asset_id = ?`, string(assetID)).Scan(&existingLibrary, &existingRecordingID, &existingSource, &existingPath, &existingSize)
	if err == nil {
		if existingLibrary != string(libraryID) || existingRecordingID != string(recordingID) || !existingSource.Valid || existingSource.String != string(sourceItemID) || existingPath != relativePath || !existingSize.Valid || existingSize.Int64 != sizeBytes {
			return false, fmt.Errorf("playable asset id %q conflicts with existing asset", assetID)
		}
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("load playable asset for ingestion: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO playable_assets(playable_asset_id, library_id, recording_id, source_item_id, media_path, media_sha256, media_size_bytes, codec, container, duration_ms, added_at) VALUES(?, ?, ?, ?, ?, NULL, ?, NULL, NULL, NULL, ?)`, string(assetID), string(libraryID), string(recordingID), string(sourceItemID), relativePath, sizeBytes, addedAt.Format(time.RFC3339Nano)); err != nil {
		return false, fmt.Errorf("create playable asset for ingestion: %w", err)
	}
	return true, nil
}

func localSourceItemID(libraryID domain.LibraryID, fileID string) (domain.SourceItemID, error) {
	sum := sha256.Sum256([]byte(string(libraryID) + "\x00" + fileID + "\x00source-item"))
	return domain.NewSourceItemID(fmt.Sprintf("source-item:%x", sum))
}

func localPlayableAssetID(libraryID domain.LibraryID, fileID string) (domain.PlayableAssetID, error) {
	sum := sha256.Sum256([]byte(string(libraryID) + "\x00" + fileID + "\x00playable-asset"))
	return domain.NewPlayableAssetID(fmt.Sprintf("playable-asset:%x", sum))
}
