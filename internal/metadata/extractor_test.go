package metadata

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

func TestExtractID3v23TextFrames(t *testing.T) {
	payload := bytes.Join([][]byte{
		id3TextFrame(3, "TIT2", "Example Song"),
		id3TextFrame(3, "TPE1", "Example Artist"),
		id3TextFrame(3, "TALB", "Example Album"),
		id3TextFrame(3, "TPE2", "Album Artist"),
		id3TextFrame(3, "TCON", "Ambient"),
		id3TextFrame(3, "TRCK", "3/12"),
		id3TextFrame(3, "TPOS", "1/2"),
		id3TextFrame(3, "TYER", "2026"),
	}, nil)
	data := id3Tag(3, payload)

	got, err := Extract(bytes.NewReader(data), "track.mp3")
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	want := Tags{
		TagFormat:   "id3v2.3",
		Title:       "Example Song",
		Artist:      "Example Artist",
		Album:       "Example Album",
		AlbumArtist: "Album Artist",
		Genre:       "Ambient",
		Date:        "2026",
		TrackNumber: "3/12",
		DiscNumber:  "1/2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tags = %#v, want %#v", got, want)
	}
}

func TestExtractID3v24UsesSynchsafeFrameSize(t *testing.T) {
	payload := bytes.Join([][]byte{
		id3TextFrame(4, "TIT2", "Version Four"),
		id3TextFrame(4, "TDRC", "2026-09-15"),
	}, nil)
	got, err := Extract(bytes.NewReader(id3Tag(4, payload)), ".mp3")
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if got.TagFormat != "id3v2.4" || got.Title != "Version Four" || got.Date != "2026-09-15" {
		t.Fatalf("tags = %#v", got)
	}
}

func TestExtractMP3WithoutID3ReturnsExplicitNone(t *testing.T) {
	got, err := Extract(bytes.NewReader([]byte{0xff, 0xfb, 0x90, 0x64}), ".mp3")
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if got.TagFormat != "none" {
		t.Fatalf("TagFormat = %q, want none", got.TagFormat)
	}
}

func TestExtractFLACVorbisComments(t *testing.T) {
	comments := []string{
		"TITLE=FLAC Song",
		"ARTIST=FLAC Artist",
		"ALBUM=FLAC Album",
		"ALBUMARTIST=FLAC Album Artist",
		"GENRE=Electronic",
		"DATE=2026-09-15",
		"TRACKNUMBER=4",
		"DISCNUMBER=2",
	}
	payload := flacVorbisPayload("GoreeCloud test", comments)
	data := append([]byte("fLaC"), byte(0x84), byte(len(payload)>>16), byte(len(payload)>>8), byte(len(payload)))
	data = append(data, payload...)

	got, err := Extract(bytes.NewReader(data), "track.flac")
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	want := Tags{
		TagFormat:   "flac-vorbis-comment",
		Title:       "FLAC Song",
		Artist:      "FLAC Artist",
		Album:       "FLAC Album",
		AlbumArtist: "FLAC Album Artist",
		Genre:       "Electronic",
		Date:        "2026-09-15",
		TrackNumber: "4",
		DiscNumber:  "2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tags = %#v, want %#v", got, want)
	}
}

func TestExtractRejectsMalformedAndUnsupportedInput(t *testing.T) {
	malformed := append([]byte{'I', 'D', '3', 3, 0, 0}, []byte{0x80, 0, 0, 1}...)
	if _, err := Extract(bytes.NewReader(malformed), ".mp3"); err == nil {
		t.Fatal("expected malformed ID3 size to fail")
	}
	if _, err := Extract(bytes.NewReader([]byte("not flac")), ".flac"); err == nil {
		t.Fatal("expected invalid FLAC signature to fail")
	}
	if _, err := Extract(bytes.NewReader([]byte("audio")), ".m4a"); err == nil {
		t.Fatal("expected unsupported metadata container to fail")
	}
}

func id3Tag(version byte, payload []byte) []byte {
	header := []byte{'I', 'D', '3', version, 0, 0}
	header = append(header, synchsafe(len(payload))...)
	return append(header, payload...)
}

func id3TextFrame(version byte, id, value string) []byte {
	data := append([]byte{3}, []byte(value)...)
	frame := []byte(id)
	if version == 4 {
		frame = append(frame, synchsafe(len(data))...)
	} else {
		size := make([]byte, 4)
		binary.BigEndian.PutUint32(size, uint32(len(data)))
		frame = append(frame, size...)
	}
	frame = append(frame, 0, 0)
	return append(frame, data...)
}

func synchsafe(value int) []byte {
	return []byte{
		byte((value >> 21) & 0x7f),
		byte((value >> 14) & 0x7f),
		byte((value >> 7) & 0x7f),
		byte(value & 0x7f),
	}
}

func flacVorbisPayload(vendor string, comments []string) []byte {
	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, uint32(len(vendor)))
	out.WriteString(vendor)
	_ = binary.Write(&out, binary.LittleEndian, uint32(len(comments)))
	for _, comment := range comments {
		_ = binary.Write(&out, binary.LittleEndian, uint32(len(comment)))
		out.WriteString(comment)
	}
	return out.Bytes()
}
