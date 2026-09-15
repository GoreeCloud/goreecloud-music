package sqlitestore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

// PlayableAssetIntegrity is a bounded full-file integrity fact for an
// already-materialized GoreeCloud Server playable asset. SHA256 identifies the
// exact source-file bytes read by this operation; it is not Recording identity.
type PlayableAssetIntegrity struct {
	PlayableAssetID domain.PlayableAssetID
	SHA256          string
}

// HashLibraryFileContent computes and persists SHA-256 over the complete,
// current source file for an already-materialized playable asset. It rechecks
// library edit authorization, scanner/source binding, rooted source safety,
// and scanner freshness. The operation does not mutate source media and does
// not use the digest as canonical Recording or Release identity.
func (b *Backend) HashLibraryFileContent(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, fileID string) (PlayableAssetIntegrity, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return PlayableAssetIntegrity{}, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return PlayableAssetIntegrity{}, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" || len(fileID) > 256 {
		return PlayableAssetIntegrity{}, fmt.Errorf("library file id must contain 1 to 256 characters")
	}
	sourceItemID, err := localSourceItemID(validatedLibrary, fileID)
	if err != nil {
		return PlayableAssetIntegrity{}, err
	}
	assetID, err := localPlayableAssetID(validatedLibrary, fileID)
	if err != nil {
		return PlayableAssetIntegrity{}, err
	}

	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return PlayableAssetIntegrity{}, err
	}
	defer func() { _ = tx.Rollback() }()

	allowed, err := txHasLibraryEditPermission(ctx, tx, validatedProfile, validatedLibrary)
	if err != nil {
		return PlayableAssetIntegrity{}, err
	}
	if !allowed {
		return PlayableAssetIntegrity{}, fmt.Errorf("library edit permission required")
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
		return PlayableAssetIntegrity{}, fmt.Errorf("materialized playable asset is required before integrity hashing")
	}
	if err != nil {
		return PlayableAssetIntegrity{}, fmt.Errorf("load playable asset for integrity hashing: %w", err)
	}
	if missingSince.Valid {
		return PlayableAssetIntegrity{}, fmt.Errorf("library file is currently missing")
	}
	if sourceRecording == "" || sourceRecording != assetRecording || mediaPath != relativePath || !mediaSize.Valid || mediaSize.Int64 != sizeBytes {
		return PlayableAssetIntegrity{}, fmt.Errorf("playable asset binding no longer matches scanner/source state")
	}

	source, err := openObservedLibraryFile(rootPath, relativePath, sizeBytes, modifiedNS)
	if err != nil {
		return PlayableAssetIntegrity{}, err
	}
	defer source.Close()

	digest, err := sha256WithContext(ctx, source)
	if err != nil {
		return PlayableAssetIntegrity{}, fmt.Errorf("hash library file %q: %w", relativePath, err)
	}
	if err := verifyObservedLibraryFileSnapshot(source, sizeBytes, modifiedNS); err != nil {
		return PlayableAssetIntegrity{}, err
	}
	digestHex := hex.EncodeToString(digest[:])

	result, err := tx.ExecContext(ctx, `UPDATE playable_assets SET media_sha256 = ?
		WHERE playable_asset_id = ? AND library_id = ? AND recording_id = ? AND source_item_id = ? AND media_path = ? AND media_size_bytes = ?`,
		digestHex, string(assetID), string(validatedLibrary), assetRecording, string(sourceItemID), mediaPath, sizeBytes)
	if err != nil {
		return PlayableAssetIntegrity{}, fmt.Errorf("persist playable asset integrity hash: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return PlayableAssetIntegrity{}, fmt.Errorf("confirm playable asset integrity update: %w", err)
	}
	if rows != 1 {
		return PlayableAssetIntegrity{}, fmt.Errorf("playable asset changed during integrity hashing")
	}
	if err := verifyObservedLibraryFileSnapshot(source, sizeBytes, modifiedNS); err != nil {
		return PlayableAssetIntegrity{}, err
	}
	if err := tx.Commit(); err != nil {
		return PlayableAssetIntegrity{}, fmt.Errorf("commit playable asset integrity hash: %w", err)
	}
	return PlayableAssetIntegrity{PlayableAssetID: assetID, SHA256: digestHex}, nil
}

func sha256WithContext(ctx context.Context, reader io.Reader) ([sha256.Size]byte, error) {
	h := sha256.New()
	buf := make([]byte, 64*1024)
	for {
		if err := ctx.Err(); err != nil {
			return [sha256.Size]byte{}, err
		}
		n, err := reader.Read(buf)
		if n > 0 {
			if _, writeErr := h.Write(buf[:n]); writeErr != nil {
				return [sha256.Size]byte{}, writeErr
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return [sha256.Size]byte{}, err
		}
		if n == 0 {
			return [sha256.Size]byte{}, io.ErrNoProgress
		}
	}
	var digest [sha256.Size]byte
	copy(digest[:], h.Sum(nil))
	return digest, nil
}
