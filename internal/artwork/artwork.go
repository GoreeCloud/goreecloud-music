package artwork

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

func Detect(ctx context.Context, audioPath string) (Source, bool, error) {
	if source, found, err := DetectSidecar(ctx, audioPath); err != nil || found {
		return source, found, err
	}
	return DetectEmbedded(ctx, audioPath)
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

func DetectEmbedded(ctx context.Context, audioPath string) (Source, bool, error) {
	audioPath = strings.TrimSpace(audioPath)
	if audioPath == "" {
		return Source{}, false, errors.New("audio path is required")
	}
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-select_streams", "v", "-show_entries", "stream=codec_name:stream_disposition=attached_pic", "-of", "json", audioPath)
	output, err := cmd.Output()
	if err != nil {
		return Source{}, false, fmt.Errorf("ffprobe artwork: %w", err)
	}
	mimeType, found, err := ParseEmbeddedProbe(output)
	if err != nil || !found {
		return Source{}, found, err
	}
	return Source{Type: SourceEmbedded, Path: audioPath, MIMEType: mimeType}, true, nil
}

func ParseEmbeddedProbe(data []byte) (string, bool, error) {
	var result struct {
		Streams []struct {
			CodecName   string `json:"codec_name"`
			Disposition struct {
				AttachedPic int `json:"attached_pic"`
			} `json:"disposition"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", false, fmt.Errorf("decode artwork probe: %w", err)
	}
	for _, stream := range result.Streams {
		if stream.Disposition.AttachedPic != 1 {
			continue
		}
		return mimeForCodec(stream.CodecName), true, nil
	}
	return "", false, nil
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

func mimeForCodec(codec string) string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "mjpeg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
