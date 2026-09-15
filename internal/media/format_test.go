package media

import (
	"bytes"
	"testing"
)

func TestProbeFormatMP3WithAndWithoutID3(t *testing.T) {
	frames := testMP3Frames()
	withID3 := append([]byte{'I', 'D', '3', 3, 0, 0, 0, 0, 0, 0}, frames...)
	for name, source := range map[string][]byte{
		"with-id3":   withID3,
		"raw-frames": frames,
	} {
		t.Run(name, func(t *testing.T) {
			got, err := ProbeFormat(bytes.NewReader(source), ".mp3")
			if err != nil {
				t.Fatalf("ProbeFormat() error = %v", err)
			}
			if got.Codec != "mp3" || got.Container != "mpeg-audio" {
				t.Fatalf("format = %#v", got)
			}
		})
	}
}

func TestProbeFormatRejectsID3WithoutMPEGAudio(t *testing.T) {
	source := []byte{'I', 'D', '3', 3, 0, 0, 0, 0, 0, 0}
	if _, err := ProbeFormat(bytes.NewReader(source), ".mp3"); err == nil {
		t.Fatal("expected tag-only MP3 probe to fail")
	}
}

func TestProbeFormatRejectsNonLayerIIIFrames(t *testing.T) {
	frame := make([]byte, 417)
	copy(frame, []byte{0xff, 0xfd, 0x90, 0x64}) // MPEG-1 Layer II, not Layer III.
	source := append(append([]byte{}, frame...), frame...)
	if _, err := ProbeFormat(bytes.NewReader(source), ".mp3"); err == nil {
		t.Fatal("expected Layer II probe to fail")
	}
}

func TestProbeFormatFLACRequiresSTREAMINFO(t *testing.T) {
	valid := append([]byte{'f', 'L', 'a', 'C', 0x80, 0, 0, 34}, make([]byte, 34)...)
	got, err := ProbeFormat(bytes.NewReader(valid), ".flac")
	if err != nil {
		t.Fatalf("ProbeFormat() error = %v", err)
	}
	if got.Codec != "flac" || got.Container != "flac" {
		t.Fatalf("format = %#v", got)
	}

	invalid := append([]byte{'f', 'L', 'a', 'C', 0x80, 0, 0, 33}, make([]byte, 33)...)
	if _, err := ProbeFormat(bytes.NewReader(invalid), ".flac"); err == nil {
		t.Fatal("expected invalid STREAMINFO length to fail")
	}
}

func TestProbeFormatRejectsUnsupportedExtension(t *testing.T) {
	if _, err := ProbeFormat(bytes.NewReader(testMP3Frames()), ".wav"); err == nil {
		t.Fatal("expected unsupported format probe to fail")
	}
}

func testMP3Frames() []byte {
	const frameLength = 417 // MPEG-1 Layer III, 128 kbps, 44.1 kHz, no padding.
	frame := make([]byte, frameLength)
	copy(frame, []byte{0xff, 0xfb, 0x90, 0x64})
	return append(append([]byte{}, frame...), frame...)
}
