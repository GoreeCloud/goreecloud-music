package scanner

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

var ErrInvalidRoot = errors.New("invalid library root")

var supportedExtensions = map[string]struct{}{
	".aac":  {},
	".aiff": {},
	".alac": {},
	".flac": {},
	".m4a":  {},
	".mp3":  {},
	".ogg":  {},
	".opus": {},
	".wav":  {},
}

type File struct {
	Path      string
	Extension string
	SizeBytes int64
}

func Discover(root string) ([]File, error) {
	cleanRoot := filepath.Clean(strings.TrimSpace(root))
	if cleanRoot == "." || cleanRoot == "" {
		return nil, ErrInvalidRoot
	}

	info, err := fs.Stat(osDirFS{}, cleanRoot)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, ErrInvalidRoot
	}

	var files []File
	err = filepath.WalkDir(cleanRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if _, ok := supportedExtensions[ext]; !ok {
			return nil
		}
		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		files = append(files, File{Path: path, Extension: ext, SizeBytes: fileInfo.Size()})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

type osDirFS struct{}

func (osDirFS) Open(name string) (fs.File, error) {
	return fs.OpenFile(nil, name)
}
