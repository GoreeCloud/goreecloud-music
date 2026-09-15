package media

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const (
	maxID3ProbeBytes = 8 << 20
	maxMP3ScanBytes  = 64 << 10
	maxMP3ReadBytes  = maxMP3ScanBytes + 4096
)

// Format contains bounded, source-verified media format facts. It deliberately
// excludes duration, bitrate, sample rate, channels, hashes, and other facts
// that require separate probing and acceptance.
type Format struct {
	Codec     string
	Container string
}

// ProbeFormat validates the scanner-approved extension against the actual
// source bytes. An extension selects the bounded parser; it never establishes
// the format by itself.
func ProbeFormat(source io.ReadSeeker, extension string) (Format, error) {
	if source == nil {
		return Format{}, fmt.Errorf("media source must not be nil")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return Format{}, fmt.Errorf("seek media source: %w", err)
	}

	switch strings.ToLower(filepath.Ext(extension)) {
	case ".mp3":
		return probeMP3(source)
	case ".flac":
		return probeFLAC(source)
	default:
		return Format{}, fmt.Errorf("media format probing is not supported for %q", extension)
	}
}

func probeFLAC(source io.Reader) (Format, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(source, header); err != nil {
		return Format{}, fmt.Errorf("read FLAC header: %w", err)
	}
	if string(header[:4]) != "fLaC" {
		return Format{}, fmt.Errorf("FLAC signature is missing")
	}
	if header[4]&0x7f != 0 {
		return Format{}, fmt.Errorf("FLAC first metadata block is not STREAMINFO")
	}
	length := int(header[5])<<16 | int(header[6])<<8 | int(header[7])
	if length != 34 {
		return Format{}, fmt.Errorf("FLAC STREAMINFO length is %d, expected 34", length)
	}
	if _, err := io.CopyN(io.Discard, source, int64(length)); err != nil {
		return Format{}, fmt.Errorf("read FLAC STREAMINFO: %w", err)
	}
	return Format{Codec: "flac", Container: "flac"}, nil
}

func probeMP3(source io.ReadSeeker) (Format, error) {
	audioOffset, err := mp3AudioOffset(source)
	if err != nil {
		return Format{}, err
	}
	if _, err := source.Seek(audioOffset, io.SeekStart); err != nil {
		return Format{}, fmt.Errorf("seek MP3 audio payload: %w", err)
	}
	data, err := io.ReadAll(io.LimitReader(source, maxMP3ReadBytes))
	if err != nil {
		return Format{}, fmt.Errorf("read MP3 probe window: %w", err)
	}
	if len(data) < 8 {
		return Format{}, fmt.Errorf("MP3 audio payload is too short")
	}

	limit := len(data) - 4
	if limit > maxMP3ScanBytes {
		limit = maxMP3ScanBytes
	}
	for offset := 0; offset <= limit; offset++ {
		first, ok := parseMP3FrameHeader(data[offset:])
		if !ok {
			continue
		}
		nextOffset := offset + first.length
		if nextOffset+4 > len(data) {
			continue
		}
		second, ok := parseMP3FrameHeader(data[nextOffset:])
		if !ok || second.version != first.version || second.sampleRate != first.sampleRate {
			continue
		}
		return Format{Codec: "mp3", Container: "mpeg-audio"}, nil
	}
	return Format{}, fmt.Errorf("no consecutive MPEG Layer III frames found")
}

func mp3AudioOffset(source io.ReadSeeker) (int64, error) {
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return 0, fmt.Errorf("seek MP3 source: %w", err)
	}
	header := make([]byte, 10)
	n, err := io.ReadFull(source, header)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			if _, seekErr := source.Seek(0, io.SeekStart); seekErr != nil {
				return 0, fmt.Errorf("rewind short MP3 source: %w", seekErr)
			}
			if n < 3 || string(header[:3]) != "ID3" {
				return 0, nil
			}
		}
		return 0, fmt.Errorf("read ID3 header: %w", err)
	}
	if string(header[:3]) != "ID3" {
		return 0, nil
	}
	version := header[3]
	if version != 3 && version != 4 {
		return 0, fmt.Errorf("unsupported ID3v2 major version %d", version)
	}
	if header[4] != 0 {
		return 0, fmt.Errorf("unsupported ID3v2 revision %d.%d", version, header[4])
	}
	tagSize, err := decodeSynchsafe32(header[6:10])
	if err != nil {
		return 0, fmt.Errorf("decode ID3 tag size: %w", err)
	}
	if tagSize > maxID3ProbeBytes {
		return 0, fmt.Errorf("ID3 tag size %d exceeds %d-byte probe limit", tagSize, maxID3ProbeBytes)
	}
	offset := int64(10 + tagSize)
	if version == 4 && header[5]&0x10 != 0 {
		offset += 10
	}
	return offset, nil
}

func decodeSynchsafe32(value []byte) (int, error) {
	if len(value) != 4 {
		return 0, fmt.Errorf("synchsafe integer must contain 4 bytes")
	}
	for _, b := range value {
		if b&0x80 != 0 {
			return 0, fmt.Errorf("synchsafe integer contains high-bit data")
		}
	}
	return int(value[0])<<21 | int(value[1])<<14 | int(value[2])<<7 | int(value[3]), nil
}

type mp3Frame struct {
	length     int
	version    byte
	sampleRate int
}

func parseMP3FrameHeader(data []byte) (mp3Frame, bool) {
	if len(data) < 4 || data[0] != 0xff || data[1]&0xe0 != 0xe0 {
		return mp3Frame{}, false
	}
	version := (data[1] >> 3) & 0x03
	if version == 0x01 {
		return mp3Frame{}, false
	}
	if (data[1]>>1)&0x03 != 0x01 { // Layer III only.
		return mp3Frame{}, false
	}
	bitrateIndex := (data[2] >> 4) & 0x0f
	if bitrateIndex == 0 || bitrateIndex == 0x0f {
		return mp3Frame{}, false
	}
	sampleIndex := (data[2] >> 2) & 0x03
	if sampleIndex == 0x03 || data[3]&0x03 == 0x02 {
		return mp3Frame{}, false
	}

	bitrate := mp3BitrateKbps(version, bitrateIndex)
	sampleRate := mp3SampleRate(version, sampleIndex)
	if bitrate == 0 || sampleRate == 0 {
		return mp3Frame{}, false
	}
	padding := int((data[2] >> 1) & 0x01)
	factor := 72000
	if version == 0x03 {
		factor = 144000
	}
	length := factor*bitrate/sampleRate + padding
	if length <= 4 {
		return mp3Frame{}, false
	}
	return mp3Frame{length: length, version: version, sampleRate: sampleRate}, true
}

func mp3BitrateKbps(version, index byte) int {
	mpeg1 := [...]int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0}
	mpeg2 := [...]int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}
	if version == 0x03 {
		return mpeg1[index]
	}
	return mpeg2[index]
}

func mp3SampleRate(version, index byte) int {
	var rates [3]int
	switch version {
	case 0x03:
		rates = [3]int{44100, 48000, 32000}
	case 0x02:
		rates = [3]int{22050, 24000, 16000}
	case 0x00:
		rates = [3]int{11025, 12000, 8000}
	default:
		return 0
	}
	return rates[index]
}
