package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var ErrNoAudioStream = errors.New("no audio stream found")

type TrackMetadata struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Genre       string
	Date        string
	Track       string
	Disc        string
	Codec       string
	Bitrate     int
	BitDepth    int
	SampleRate  int
	Channels    int
	DurationMS  int64
}

type ffprobeOutput struct {
	Format struct {
		Duration string            `json:"duration"`
		BitRate  string            `json:"bit_rate"`
		Tags     map[string]string `json:"tags"`
	} `json:"format"`
	Streams []struct {
		CodecType        string            `json:"codec_type"`
		CodecName        string            `json:"codec_name"`
		SampleRate       string            `json:"sample_rate"`
		Channels         int               `json:"channels"`
		BitRate          string            `json:"bit_rate"`
		BitsPerSample    int               `json:"bits_per_sample"`
		BitsPerRawSample string            `json:"bits_per_raw_sample"`
		Tags             map[string]string `json:"tags"`
	} `json:"streams"`
}

func Probe(ctx context.Context, path string) (TrackMetadata, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return TrackMetadata{}, errors.New("metadata path is required")
	}
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration,bit_rate:format_tags:stream=codec_type,codec_name,sample_rate,channels,bit_rate,bits_per_sample,bits_per_raw_sample:stream_tags", "-of", "json", path)
	output, err := cmd.Output()
	if err != nil {
		return TrackMetadata{}, fmt.Errorf("ffprobe: %w", err)
	}
	return ParseFFProbe(output)
}

func ParseFFProbe(data []byte) (TrackMetadata, error) {
	var result ffprobeOutput
	if err := json.Unmarshal(data, &result); err != nil {
		return TrackMetadata{}, fmt.Errorf("decode ffprobe output: %w", err)
	}

	var audio *struct {
		CodecType        string            `json:"codec_type"`
		CodecName        string            `json:"codec_name"`
		SampleRate       string            `json:"sample_rate"`
		Channels         int               `json:"channels"`
		BitRate          string            `json:"bit_rate"`
		BitsPerSample    int               `json:"bits_per_sample"`
		BitsPerRawSample string            `json:"bits_per_raw_sample"`
		Tags             map[string]string `json:"tags"`
	}
	for i := range result.Streams {
		if result.Streams[i].CodecType == "audio" {
			audio = &result.Streams[i]
			break
		}
	}
	if audio == nil {
		return TrackMetadata{}, ErrNoAudioStream
	}

	tags := mergeTags(result.Format.Tags, audio.Tags)
	bitrate := parseInt(audio.BitRate)
	if bitrate == 0 {
		bitrate = parseInt(result.Format.BitRate)
	}
	bitDepth := audio.BitsPerSample
	if raw := parseInt(audio.BitsPerRawSample); raw > bitDepth {
		bitDepth = raw
	}
	return TrackMetadata{
		Title:       tag(tags, "title"),
		Artist:      tag(tags, "artist"),
		Album:       tag(tags, "album"),
		AlbumArtist: tag(tags, "album_artist", "albumartist"),
		Genre:       tag(tags, "genre"),
		Date:        tag(tags, "date", "year"),
		Track:       tag(tags, "track"),
		Disc:        tag(tags, "disc"),
		Codec:       audio.CodecName,
		Bitrate:     bitrate,
		BitDepth:    bitDepth,
		SampleRate:  parseInt(audio.SampleRate),
		Channels:    audio.Channels,
		DurationMS:  parseDurationMS(result.Format.Duration),
	}, nil
}

func mergeTags(tagSets ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, tags := range tagSets {
		for key, value := range tags {
			merged[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
		}
	}
	return merged
}

func tag(tags map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(tags[strings.ToLower(key)]); value != "" {
			return value
		}
	}
	return ""
}

func parseInt(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func parseDurationMS(value string) int64 {
	seconds, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return int64(seconds * 1000)
}
