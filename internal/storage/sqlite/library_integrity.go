package sqlitestore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

// PlayableAssetContentHash is the bounded content-hash result persisted for an
// already-materialized GoreeCloud Server playable asset. It is source evidence,
// not production integrity qualification or a substitute for scanner state.
type PlayableAssetContentHash struct {
	PlayableAssetID domain.PlayableAssetID
	SHA256          string
	SizeBytes       int64
}

// HashLibraryFileContent computes and persists a stable SHA-256 for the exact
// currently observed source bytes of an already-materialized playable asset.
// The source is hashed twice through the same rooted, read-only file handle and
// both passes must agree. Library edit permission and source bindings are
// re-evaluated inside the transaction. Existing conflicting hashes fail closed.
func (b *Backend) HashLibraryFileContent(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, fileID string) (PlayableAssetContentHash, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return PlayableAssetContentHash{}, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return PlayableAssetContentHash{}, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" || len(fileID) > 256 {
		return PlayableAssetContentHash{}, fmt.Errorf("library file id must contain 1 to 256 characters")
	}
	sourceItemID, err := localSourceItemID(validatedLibrary, fileID)
	if err != nil {
		return PlayableAssetContentHash{}, err
	}
	assetID, err := localPlayableAssetID(validatedLibrary, fileID)
	if err != nil {
		return PlayableAssetContentHash{}, err
	}

	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return PlayableAssetContentHash{}, err
	}
	defer func() { _ = tx.Rollback() }()

	allowed, err := txHasLibraryEditPermission(ctx, tx, validatedProfile, validatedLibrary)
	if err != nil {
		return PlayableAssetContentHash{}, err
	}
	if !allowed {
		return PlayableAssetContentHash{}, fmt.Errorf("library edit permission required")
	}

	var rootPath, relativePath string
	var sizeBytes, modifiedNS int64
	var missingSince sql.NullString
	var sourceRecording, assetRecording, mediaPath string
	var mediaSize sql.NullInt64
	var existingHash sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT l.root_path, f.relative_path, f.size_bytes, f.modified_ns, f.missing_since,
		s.recording_id, a.recording_id, a.media_path, a.media_size_bytes, a.media_sha256
		FROM library_files f
		JOIN libraries l ON l.library_id = f.library_id
		JOIN source_items s ON s.source_item_id = ? AND s.library_id = f.library_id AND s.source_kind = ? AND s.external_id = f.file_id
		JOIN playable_assets a ON a.playable_asset_id = ? AND a.library_id = f.library_id AND a.source_item_id = s.source_item_id
		WHERE f.file_id = ? AND f.library_id = ?`,
		string(sourceItemID), string(domain.SourceGoreeCloudServer), string(assetID), fileID, string(validatedLibrary)).Scan(
		&rootPath, &relativePath, &sizeBytes, &modifiedNS, &missingSince,
		&sourceRecording, &assetRecording, &mediaPath, &mediaSize, &existingHash)
	if errors.Is(err, sql.ErrNoRows) {
		return PlayableAssetContentHash{}, fmt.Errorf("materialized playable asset is required before content hashing")
	}
	if err != nil {
		return PlayableAssetContentHash{}, fmt.Errorf("load playable asset for content hashing: %w", err)
	}
	if missingSince.Valid {
		return PlayableAssetContentHash{}, fmt.Errorf("library file is currently missing")
	}
	if sourceRecording == "" || sourceRecording != assetRecording || mediaPath != relativePath || !mediaSize.Valid || mediaSize.Int64 != sizeBytes {
		return PlayableAssetContentHash{}, fmt.Errorf("playable asset binding no longer matches scanner/source state")
	}

	source, err := openObservedLibraryFile(rootPath, relativePath, sizeBytes, modifiedNS)
	if err != nil {
		return PlayableAssetContentHash{}, err
	}
	defer source.Close()

	digest, err := stableSHA256(source, sizeBytes, modifiedNS)
	if err != nil {
		return PlayableAssetContentHash{}, err
	}
	encoded := hex.EncodeToString(digest)
	if existingHash.Valid {
		if existingHash.String != encoded {
			return PlayableAssetContentHash{}, fmt.Errorf("stored playable asset content hash conflicts with current source bytes; rescan and reconcile required")
		}
		if err := verifyObservedLibraryFileSnapshot(source, sizeBytes, modifiedNS); err != nil {
			return PlayableAssetContentHash{}, err
		}
		if err := tx.Commit(); err != nil {
			return PlayableAssetContentHash{}, fmt.Errorf("commit idempotent playable asset content hash verification: %w", err)
		}
		return PlayableAssetContentHash{PlayableAssetID: assetID, SHA256: encoded, SizeBytes: sizeBytes}, nil
	}

	result, err := tx.ExecContext(ctx, `UPDATE playable_assets SET media_sha256 = ?
		WHERE playable_asset_id = ? AND library_id = ? AND recording_id = ? AND source_item_id = ? AND media_path = ? AND media_size_bytes = ? AND media_sha256 IS NULL`,
		encoded, string(assetID), string(validatedLibrary), assetRecording, string(sourceItemID), mediaPath, sizeBytes)
	if err != nil {
		return PlayableAssetContentHash{}, fmt.Errorf("persist playable asset content hash: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return PlayableAssetContentHash{}, fmt.Errorf("confirm playable asset content hash update: %w", err)
	}
	if rows != 1 {
		return PlayableAssetContentHash{}, fmt.Errorf("playable asset changed during content hashing")
	}
	if err := verifyObservedLibraryFileSnapshot(source, sizeBytes, modifiedNS); err != nil {
		return PlayableAssetContentHash{}, err
	}
	if err := tx.Commit(); err != nil {
		return PlayableAssetContentHash{}, fmt.Errorf("commit playable asset content hash: %w", err)
	}
	return PlayableAssetContentHash{PlayableAssetID: assetID, SHA256: encoded, SizeBytes: sizeBytes}, nil
}

func stableSHA256(source *os.File, expectedSize, expectedModifiedNS int64) ([]byte, error) {
	var first []byte
	for pass := 0; pass < 2; pass++ {
		if _, err := source.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek library file for content hash pass %d: %w", pass+1, err)
		}
		h := sha256.New()
		written, err := io.Copy(h, source)
		if err != nil {
			return nil, fmt.Errorf("hash library file pass %d: %w", pass+1, err)
		}
		if written != expectedSize {
			return nil, fmt.Errorf("library file size changed during content hash pass %d; rescan required", pass+1)
		}
		if err := verifyObservedLibraryFileSnapshot(source, expectedSize, expectedModifiedNS); err != nil {
			return nil, err
		}
		digest := h.Sum(nil)
		if pass == 0 {
			first = append([]byte(nil), digest...)
			continue
		}
		if !bytes.Equal(first, digest) {
			return nil, fmt.Errorf("library file content changed between hash passes; rescan required")
		}
	}
	return first, nil
}
