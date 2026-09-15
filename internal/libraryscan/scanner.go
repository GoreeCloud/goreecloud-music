package libraryscan

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Observation is the minimum filesystem state needed to reconcile one audio
// file without reading or mutating its media contents.
type Observation struct {
	RelativePath string
	SizeBytes    int64
	ModifiedNS   int64
}

var supportedExtensions = map[string]struct{}{
	".aac":  {},
	".aif":  {},
	".aiff": {},
	".flac": {},
	".m4a":  {},
	".mp3":  {},
	".oga":  {},
	".ogg":  {},
	".opus": {},
	".wav":  {},
}

// Scan discovers supported regular audio files beneath root. It never follows
// symbolic links and never mutates source files. Returned paths are relative to
// root, slash-normalized, and deterministically ordered.
func Scan(root string) ([]Observation, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("library root must not be empty")
	}
	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("library root must be absolute")
	}
	root = filepath.Clean(root)
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return nil, fmt.Errorf("inspect library root: %w", err)
	}
	if rootInfo.Mode()&fs.ModeSymlink != 0 {
		return nil, fmt.Errorf("library root must not be a symbolic link")
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("library root must be a directory")
	}

	observations := make([]Observation, 0)
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if _, ok := supportedExtensions[strings.ToLower(filepath.Ext(entry.Name()))]; !ok {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("resolve path relative to library root: %w", err)
		}
		relative = filepath.ToSlash(filepath.Clean(relative))
		if relative == "." || relative == ".." || strings.HasPrefix(relative, "../") || strings.HasPrefix(relative, "/") {
			return fmt.Errorf("discovered path escapes library root: %q", relative)
		}
		observations = append(observations, Observation{
			RelativePath: relative,
			SizeBytes:    info.Size(),
			ModifiedNS:   info.ModTime().UnixNano(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan library root: %w", err)
	}

	sort.Slice(observations, func(i, j int) bool {
		return observations[i].RelativePath < observations[j].RelativePath
	})
	return observations, nil
}
