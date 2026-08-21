package metadata

import (
	"errors"
	"testing"
)

func TestParseFFProbe(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"format": {
			"duration": "245.125",
			"bit_rate": "921600",
			"tags": {
				"TITLE": "Signal",
				"ARTIST": "Example Artist",
				"ALBUM": "Example Album",
				"ALBUMARTIST": "Example Artist",
				"GENRE": "Electronic",
				"DATE": "2026",
				"TRACK": "4/10",
				"DISC": "1/1"
			}
		},
		"streams": [{
			"codec_type": "audio",
			"codec_name": "flac",
			"sample_rate": "96000",
			"channels": 2,
			"bits_per_raw_sample": "24"
		}]
	}`)

	metadata, err := ParseFFProbe(input)
	if err != nil {
		t.Fatalf("parse metadata: %v", err)
	}
	if metadata.Title != "Signal" || metadata.Artist != "Example Artist" {
		t.Fatalf("unexpected tags: %+v", metadata)
	}
	if metadata.Codec != "flac" || metadata.SampleRate != 96000 || metadata.Channels != 2 || metadata.BitDepth != 24 {
		t.Fatalf("unexpected audio properties: %+v", metadata)
	}
	if metadata.Bitrate != 921600 || metadata.DurationMS != 245125 {
		t.Fatalf("unexpected technical metadata: %+v", metadata)
	}
}

func TestParseFFProbeRequiresAudioStream(t *testing.T) {
	t.Parallel()

	_, err := ParseFFProbe([]byte(`{"format":{},"streams":[]}`))
	if !errors.Is(err, ErrNoAudioStream) {
		t.Fatalf("expected ErrNoAudioStream, got %v", err)
	}
}
