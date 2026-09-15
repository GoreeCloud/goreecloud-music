package sqlitestore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/libraryscan"
)

func init() {
	entityDDL["library_files"] = `CREATE TABLE IF NOT EXISTS library_files (
		file_id TEXT PRIMARY KEY,
		library_id TEXT NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
		relative_path TEXT NOT NULL,
		size_bytes INTEGER NOT NULL CHECK(size_bytes >= 0),
		modified_ns INTEGER NOT NULL,
		first_seen_at TEXT NOT NULL,
		last_seen_at TEXT NOT NULL,
		missing_since TEXT,
		UNIQUE(library_id, relative_path)
	) STRICT`
}

// LibraryFile is one durable filesystem observation. It stores only file facts
// and references; source media bytes remain in the library filesystem.
type LibraryFile struct {
	ID           string
	LibraryID    domain.LibraryID
	RelativePath string
	SizeBytes    int64
	ModifiedNS   int64
	FirstSeenAt  time.Time
	LastSeenAt   time.Time
	MissingSince *time.Time
}

// ReconcileResult describes a single library reconciliation without claiming
// metadata ingestion or canonical recording creation.
type ReconcileResult struct {
	Added     int
	Updated   int
	Unchanged int
	Missing   int
	Restored  int
}

// ScanLibrary discovers supported source files beneath the configured library
// root and reconciles their filesystem facts. Edit permission is required.
func (b *Backend) ScanLibrary(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, observedAt time.Time) (ReconcileResult, error) {
	allowed, err := b.HasLibraryPermission(ctx, profileID, libraryID, LibraryPermissionEdit)
	if err != nil {
		return ReconcileResult{}, err
	}
	if !allowed {
		return ReconcileResult{}, fmt.Errorf("library edit permission required")
	}

	var rootPath string
	if err := b.db.QueryRowContext(ctx, `SELECT root_path FROM libraries WHERE library_id = ?`, string(libraryID)).Scan(&rootPath); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ReconcileResult{}, fmt.Errorf("library not found")
		}
		return ReconcileResult{}, fmt.Errorf("load library root: %w", err)
	}
	observations, err := libraryscan.Scan(rootPath)
	if err != nil {
		return ReconcileResult{}, err
	}
	return b.ReconcileLibraryFiles(ctx, profileID, libraryID, observations, observedAt)
}

// ReconcileLibraryFiles applies a complete observation set for one library.
// Files absent from the observation are tombstoned as missing rather than
// deleted, preserving identity for later restoration or explicit cleanup.
func (b *Backend) ReconcileLibraryFiles(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, observations []libraryscan.Observation, observedAt time.Time) (ReconcileResult, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return ReconcileResult{}, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return ReconcileResult{}, err
	}
	allowed, err := b.HasLibraryPermission(ctx, validatedProfile, validatedLibrary, LibraryPermissionEdit)
	if err != nil {
		return ReconcileResult{}, err
	}
	if !allowed {
		return ReconcileResult{}, fmt.Errorf("library edit permission required")
	}
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	} else {
		observedAt = observedAt.UTC()
	}

	normalized := make([]libraryscan.Observation, 0, len(observations))
	seenInput := make(map[string]struct{}, len(observations))
	for _, observation := range observations {
		path, err := normalizeLibraryRelativePath(observation.RelativePath)
		if err != nil {
			return ReconcileResult{}, err
		}
		if observation.SizeBytes < 0 {
			return ReconcileResult{}, fmt.Errorf("library file %q has negative size", path)
		}
		if _, duplicate := seenInput[path]; duplicate {
			return ReconcileResult{}, fmt.Errorf("duplicate library file observation %q", path)
		}
		seenInput[path] = struct{}{}
		observation.RelativePath = path
		normalized = append(normalized, observation)
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].RelativePath < normalized[j].RelativePath })

	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return ReconcileResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	type existingFile struct {
		ID           string
		SizeBytes    int64
		ModifiedNS   int64
		MissingSince sql.NullString
	}
	existing := make(map[string]existingFile)
	rows, err := tx.QueryContext(ctx, `SELECT file_id, relative_path, size_bytes, modified_ns, missing_since FROM library_files WHERE library_id = ?`, string(validatedLibrary))
	if err != nil {
		return ReconcileResult{}, fmt.Errorf("load library files: %w", err)
	}
	for rows.Next() {
		var file existingFile
		var relative string
		if err := rows.Scan(&file.ID, &relative, &file.SizeBytes, &file.ModifiedNS, &file.MissingSince); err != nil {
			_ = rows.Close()
			return ReconcileResult{}, err
		}
		existing[relative] = file
	}
	if err := rows.Close(); err != nil {
		return ReconcileResult{}, err
	}
	if err := rows.Err(); err != nil {
		return ReconcileResult{}, err
	}

	result := ReconcileResult{}
	seen := make(map[string]struct{}, len(normalized))
	now := observedAt.Format(time.RFC3339Nano)
	for _, observation := range normalized {
		seen[observation.RelativePath] = struct{}{}
		file, found := existing[observation.RelativePath]
		if !found {
			_, err := tx.ExecContext(ctx, `INSERT INTO library_files(file_id, library_id, relative_path, size_bytes, modified_ns, first_seen_at, last_seen_at, missing_since)
				VALUES(?, ?, ?, ?, ?, ?, ?, NULL)`, libraryFileID(validatedLibrary, observation.RelativePath), string(validatedLibrary), observation.RelativePath, observation.SizeBytes, observation.ModifiedNS, now, now)
			if err != nil {
				return ReconcileResult{}, fmt.Errorf("insert library file %q: %w", observation.RelativePath, err)
			}
			result.Added++
			continue
		}

		restored := file.MissingSince.Valid
		changed := file.SizeBytes != observation.SizeBytes || file.ModifiedNS != observation.ModifiedNS
		_, err := tx.ExecContext(ctx, `UPDATE library_files SET size_bytes = ?, modified_ns = ?, last_seen_at = ?, missing_since = NULL WHERE file_id = ?`, observation.SizeBytes, observation.ModifiedNS, now, file.ID)
		if err != nil {
			return ReconcileResult{}, fmt.Errorf("update library file %q: %w", observation.RelativePath, err)
		}
		switch {
		case restored:
			result.Restored++
		case changed:
			result.Updated++
		default:
			result.Unchanged++
		}
	}

	for relative, file := range existing {
		if _, ok := seen[relative]; ok || file.MissingSince.Valid {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE library_files SET missing_since = ? WHERE file_id = ?`, now, file.ID); err != nil {
			return ReconcileResult{}, fmt.Errorf("mark library file %q missing: %w", relative, err)
		}
		result.Missing++
	}

	if err := tx.Commit(); err != nil {
		return ReconcileResult{}, fmt.Errorf("commit library file reconciliation: %w", err)
	}
	return result, nil
}

// LibraryFilesForProfile returns only files from a library the profile can
// read. Missing tombstones are excluded unless explicitly requested.
func (b *Backend) LibraryFilesForProfile(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, includeMissing bool) ([]LibraryFile, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return nil, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return nil, err
	}
	allowed, err := b.HasLibraryPermission(ctx, validatedProfile, validatedLibrary, LibraryPermissionRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("library read permission required")
	}

	query := `SELECT file_id, relative_path, size_bytes, modified_ns, first_seen_at, last_seen_at, missing_since FROM library_files WHERE library_id = ?`
	if !includeMissing {
		query += ` AND missing_since IS NULL`
	}
	query += ` ORDER BY relative_path`
	rows, err := b.db.QueryContext(ctx, query, string(validatedLibrary))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]LibraryFile, 0)
	for rows.Next() {
		var fileID, relative, firstSeenRaw, lastSeenRaw string
		var sizeBytes, modifiedNS int64
		var missingRaw sql.NullString
		if err := rows.Scan(&fileID, &relative, &sizeBytes, &modifiedNS, &firstSeenRaw, &lastSeenRaw, &missingRaw); err != nil {
			return nil, err
		}
		firstSeen, err := time.Parse(time.RFC3339Nano, firstSeenRaw)
		if err != nil {
			return nil, fmt.Errorf("parse library file first_seen_at: %w", err)
		}
		lastSeen, err := time.Parse(time.RFC3339Nano, lastSeenRaw)
		if err != nil {
			return nil, fmt.Errorf("parse library file last_seen_at: %w", err)
		}
		file := LibraryFile{
			ID:           fileID,
			LibraryID:    validatedLibrary,
			RelativePath: relative,
			SizeBytes:    sizeBytes,
			ModifiedNS:   modifiedNS,
			FirstSeenAt:  firstSeen,
			LastSeenAt:   lastSeen,
		}
		if missingRaw.Valid {
			missingAt, err := time.Parse(time.RFC3339Nano, missingRaw.String)
			if err != nil {
				return nil, fmt.Errorf("parse library file missing_since: %w", err)
			}
			file.MissingSince = &missingAt
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func normalizeLibraryRelativePath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("library file relative path must not be empty")
	}
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("library file path must be relative: %q", value)
	}
	cleaned := filepath.ToSlash(filepath.Clean(value))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("library file path escapes library root: %q", value)
	}
	return cleaned, nil
}

func libraryFileID(libraryID domain.LibraryID, relativePath string) string {
	sum := sha256.Sum256([]byte(string(libraryID) + "\x00" + relativePath))
	return fmt.Sprintf("library-file:%x", sum)
}
