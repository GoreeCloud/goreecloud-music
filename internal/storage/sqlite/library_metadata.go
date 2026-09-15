package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/metadata"
)

func init() {
	entityDDL["library_file_metadata"] = `CREATE TABLE IF NOT EXISTS library_file_metadata (
		file_id TEXT PRIMARY KEY REFERENCES library_files(file_id) ON DELETE CASCADE,
		library_id TEXT NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
		tag_format TEXT NOT NULL,
		title TEXT NOT NULL DEFAULT '',
		artist TEXT NOT NULL DEFAULT '',
		album TEXT NOT NULL DEFAULT '',
		album_artist TEXT NOT NULL DEFAULT '',
		genre TEXT NOT NULL DEFAULT '',
		date_text TEXT NOT NULL DEFAULT '',
		track_number TEXT NOT NULL DEFAULT '',
		disc_number TEXT NOT NULL DEFAULT '',
		source_size_bytes INTEGER NOT NULL CHECK(source_size_bytes >= 0),
		source_modified_ns INTEGER NOT NULL,
		extracted_at TEXT NOT NULL
	) STRICT`
}

// LibraryFileMetadata is normalized embedded metadata bound to the exact
// scanner observation from which it was extracted. Current is false when the
// scanner facts no longer match the extraction snapshot or the file is missing.
type LibraryFileMetadata struct {
	FileID           string
	LibraryID        domain.LibraryID
	RelativePath     string
	TagFormat        string
	Title            string
	Artist           string
	Album            string
	AlbumArtist      string
	Genre            string
	Date             string
	TrackNumber      string
	DiscNumber       string
	SourceSizeBytes  int64
	SourceModifiedNS int64
	ExtractedAt      time.Time
	Current          bool
}

// ExtractLibraryFileMetadata reads embedded metadata from one currently
// observed source file and stores only normalized textual facts. It never
// writes source media and it does not create canonical Recording/Release
// identities. Edit permission is required because the operation mutates
// durable library application state.
func (b *Backend) ExtractLibraryFileMetadata(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, fileID string, extractedAt time.Time) (LibraryFileMetadata, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return LibraryFileMetadata{}, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return LibraryFileMetadata{}, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" || len(fileID) > 256 {
		return LibraryFileMetadata{}, fmt.Errorf("library file id must contain 1 to 256 characters")
	}
	allowed, err := b.HasLibraryPermission(ctx, validatedProfile, validatedLibrary, LibraryPermissionEdit)
	if err != nil {
		return LibraryFileMetadata{}, err
	}
	if !allowed {
		return LibraryFileMetadata{}, fmt.Errorf("library edit permission required")
	}

	var rootPath, relativePath string
	var sizeBytes, modifiedNS int64
	var missingSince sql.NullString
	err = b.db.QueryRowContext(ctx, `SELECT l.root_path, f.relative_path, f.size_bytes, f.modified_ns, f.missing_since
		FROM library_files f JOIN libraries l ON l.library_id = f.library_id
		WHERE f.file_id = ? AND f.library_id = ?`, fileID, string(validatedLibrary)).Scan(&rootPath, &relativePath, &sizeBytes, &modifiedNS, &missingSince)
	if errors.Is(err, sql.ErrNoRows) {
		return LibraryFileMetadata{}, fmt.Errorf("library file not found")
	}
	if err != nil {
		return LibraryFileMetadata{}, fmt.Errorf("load library file observation: %w", err)
	}
	if missingSince.Valid {
		return LibraryFileMetadata{}, fmt.Errorf("library file is currently missing")
	}
	relativePath, err = normalizeLibraryRelativePath(relativePath)
	if err != nil {
		return LibraryFileMetadata{}, err
	}

	source, err := openObservedLibraryFile(rootPath, relativePath, sizeBytes, modifiedNS)
	if err != nil {
		return LibraryFileMetadata{}, err
	}
	defer source.Close()

	tags, err := metadata.Extract(source, filepath.Ext(relativePath))
	if err != nil {
		return LibraryFileMetadata{}, fmt.Errorf("extract metadata from %q: %w", relativePath, err)
	}
	if extractedAt.IsZero() {
		extractedAt = time.Now().UTC()
	} else {
		extractedAt = extractedAt.UTC()
	}

	_, err = b.db.ExecContext(ctx, `INSERT INTO library_file_metadata(
		file_id, library_id, tag_format, title, artist, album, album_artist, genre, date_text, track_number, disc_number,
		source_size_bytes, source_modified_ns, extracted_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_id) DO UPDATE SET
			library_id = excluded.library_id,
			tag_format = excluded.tag_format,
			title = excluded.title,
			artist = excluded.artist,
			album = excluded.album,
			album_artist = excluded.album_artist,
			genre = excluded.genre,
			date_text = excluded.date_text,
			track_number = excluded.track_number,
			disc_number = excluded.disc_number,
			source_size_bytes = excluded.source_size_bytes,
			source_modified_ns = excluded.source_modified_ns,
			extracted_at = excluded.extracted_at`,
		fileID, string(validatedLibrary), tags.TagFormat, tags.Title, tags.Artist, tags.Album, tags.AlbumArtist, tags.Genre,
		tags.Date, tags.TrackNumber, tags.DiscNumber, sizeBytes, modifiedNS, extractedAt.Format(time.RFC3339Nano))
	if err != nil {
		return LibraryFileMetadata{}, fmt.Errorf("store library file metadata: %w", err)
	}

	return LibraryFileMetadata{
		FileID:           fileID,
		LibraryID:        validatedLibrary,
		RelativePath:     relativePath,
		TagFormat:        tags.TagFormat,
		Title:            tags.Title,
		Artist:           tags.Artist,
		Album:            tags.Album,
		AlbumArtist:      tags.AlbumArtist,
		Genre:            tags.Genre,
		Date:             tags.Date,
		TrackNumber:      tags.TrackNumber,
		DiscNumber:       tags.DiscNumber,
		SourceSizeBytes:  sizeBytes,
		SourceModifiedNS: modifiedNS,
		ExtractedAt:      extractedAt,
		Current:          true,
	}, nil
}

// LibraryFileMetadataForProfile returns extracted metadata only when the
// profile can read the owning library. The boolean is false when no extraction
// record exists. Current reports whether its source snapshot still matches the
// latest scanner observation and that observation is not tombstoned.
func (b *Backend) LibraryFileMetadataForProfile(ctx context.Context, profileID domain.ProfileID, libraryID domain.LibraryID, fileID string) (LibraryFileMetadata, bool, error) {
	validatedProfile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return LibraryFileMetadata{}, false, err
	}
	validatedLibrary, err := domain.NewLibraryID(string(libraryID))
	if err != nil {
		return LibraryFileMetadata{}, false, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" || len(fileID) > 256 {
		return LibraryFileMetadata{}, false, fmt.Errorf("library file id must contain 1 to 256 characters")
	}
	allowed, err := b.HasLibraryPermission(ctx, validatedProfile, validatedLibrary, LibraryPermissionRead)
	if err != nil {
		return LibraryFileMetadata{}, false, err
	}
	if !allowed {
		return LibraryFileMetadata{}, false, fmt.Errorf("library read permission required")
	}

	var result LibraryFileMetadata
	var libraryRaw, extractedRaw string
	var currentSize, currentModified int64
	var missingSince sql.NullString
	err = b.db.QueryRowContext(ctx, `SELECT m.file_id, m.library_id, f.relative_path, m.tag_format, m.title, m.artist, m.album, m.album_artist,
		m.genre, m.date_text, m.track_number, m.disc_number, m.source_size_bytes, m.source_modified_ns, m.extracted_at,
		f.size_bytes, f.modified_ns, f.missing_since
		FROM library_file_metadata m JOIN library_files f ON f.file_id = m.file_id
		WHERE m.file_id = ? AND m.library_id = ? AND f.library_id = ?`, fileID, string(validatedLibrary), string(validatedLibrary)).Scan(
		&result.FileID, &libraryRaw, &result.RelativePath, &result.TagFormat, &result.Title, &result.Artist, &result.Album, &result.AlbumArtist,
		&result.Genre, &result.Date, &result.TrackNumber, &result.DiscNumber, &result.SourceSizeBytes, &result.SourceModifiedNS, &extractedRaw,
		&currentSize, &currentModified, &missingSince)
	if errors.Is(err, sql.ErrNoRows) {
		return LibraryFileMetadata{}, false, nil
	}
	if err != nil {
		return LibraryFileMetadata{}, false, fmt.Errorf("load library file metadata: %w", err)
	}
	extractedAt, err := time.Parse(time.RFC3339Nano, extractedRaw)
	if err != nil {
		return LibraryFileMetadata{}, false, fmt.Errorf("parse metadata extracted_at: %w", err)
	}
	result.LibraryID = domain.LibraryID(libraryRaw)
	result.ExtractedAt = extractedAt
	result.Current = !missingSince.Valid && result.SourceSizeBytes == currentSize && result.SourceModifiedNS == currentModified
	return result, true, nil
}

func openObservedLibraryFile(rootPath, relativePath string, expectedSize, expectedModifiedNS int64) (*os.File, error) {
	rootPath = filepath.Clean(strings.TrimSpace(rootPath))
	if !filepath.IsAbs(rootPath) {
		return nil, fmt.Errorf("library root path must be absolute")
	}
	rootInfo, err := os.Lstat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("inspect library root: %w", err)
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return nil, fmt.Errorf("library root must remain a non-symlink directory")
	}

	normalized, err := normalizeLibraryRelativePath(relativePath)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, fmt.Errorf("open library root: %w", err)
	}
	defer root.Close()

	components := strings.Split(normalized, "/")
	current := ""
	for i, component := range components {
		if component == "" || component == "." || component == ".." {
			return nil, fmt.Errorf("library file path contains invalid component")
		}
		if current == "" {
			current = component
		} else {
			current = filepath.Join(current, component)
		}
		info, err := root.Lstat(current)
		if err != nil {
			return nil, fmt.Errorf("inspect library file component %q: %w", component, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("library file path contains symbolic link component %q", component)
		}
		if i < len(components)-1 && !info.IsDir() {
			return nil, fmt.Errorf("library file parent component %q is not a directory", component)
		}
	}

	file, err := root.Open(filepath.FromSlash(normalized))
	if err != nil {
		return nil, fmt.Errorf("open library file within root: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat opened library file: %w", err)
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, fmt.Errorf("library file is not regular")
	}
	if info.Size() != expectedSize || info.ModTime().UnixNano() != expectedModifiedNS {
		_ = file.Close()
		return nil, fmt.Errorf("library file changed since last scanner observation; rescan required")
	}
	return file, nil
}
