package artwork

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type SourceType string

const (
	SourceSidecar  SourceType = "sidecar"
	SourceEmbedded SourceType = "embedded"
)

type Source struct {
	Type     SourceType
	Path     string
	MIMEType string
}

var sidecarNames = []string{
	"cover.jpg", "cover.jpeg", "cover.png", "cover.webp",
	"folder.jpg", "folder.jpeg", "folder.png", "folder.webp",
	"front.jpg", "front.jpeg", "front.png", "front.webp",
}

func DetectSidecar(_ context.Context, audioPath string) (Source, bool, error) {
	audioPath = strings.TrimSpace(audioPath)
	if audioPath == "" {
		return Source{}, false, errors.New("audio path is required")
	}

	dir := filepath.Dir(audioPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Source{}, false, err
	}
	byLowerName := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		byLowerName[strings.ToLower(entry.Name())] = entry.Name()
	}

	for _, candidate := range sidecarNames {
		actual, ok := byLowerName[candidate]
		if !ok {
			continue
		}
		path := filepath.Join(dir, actual)
		return Source{Type: SourceSidecar, Path: path, MIMEType: mimeForPath(path)}, true, nil
	}
	return Source{}, false, nil
}

func mimeForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
