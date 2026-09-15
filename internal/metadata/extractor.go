package metadata

import (
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	maxTagBytes     = 8 << 20
	maxTextBytes    = 64 << 10
	maxVorbisFields = 10000
)

// Tags is normalized embedded metadata extracted from a source file. It is
// deliberately limited to identity-oriented textual fields needed by the next
// ingestion stage; it does not claim canonical Recording or Release identity.
type Tags struct {
	TagFormat   string
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Genre       string
	Date        string
	TrackNumber string
	DiscNumber  string
}

// Extract reads embedded metadata from the supplied seekable source. extension
// is the scanner-approved file extension and is used only to select the parser;
// each parser still validates its own container/tag signature. The source is
// never written.
func Extract(source io.ReadSeeker, extension string) (Tags, error) {
	if source == nil {
		return Tags{}, fmt.Errorf("metadata source must not be nil")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return Tags{}, fmt.Errorf("seek metadata source: %w", err)
	}

	switch strings.ToLower(filepath.Ext(extension)) {
	case ".mp3":
		return extractID3v2(source)
	case ".flac":
		return extractFLAC(source)
	default:
		return Tags{}, fmt.Errorf("embedded metadata extraction is not supported for %q", extension)
	}
}

func extractID3v2(source io.Reader) (Tags, error) {
	header := make([]byte, 10)
	n, err := io.ReadFull(source, header)
	if err != nil {
		if (err == io.EOF || err == io.ErrUnexpectedEOF) && n >= 0 {
			return Tags{TagFormat: "none"}, nil
		}
		return Tags{}, fmt.Errorf("read ID3 header: %w", err)
	}
	if string(header[:3]) != "ID3" {
		return Tags{TagFormat: "none"}, nil
	}
	version := header[3]
	if version != 3 && version != 4 {
		return Tags{}, fmt.Errorf("unsupported ID3v2 major version %d", version)
	}
	if header[4] != 0 {
		return Tags{}, fmt.Errorf("unsupported ID3v2 revision %d.%d", version, header[4])
	}
	tagSize, err := decodeSynchsafe32(header[6:10])
	if err != nil {
		return Tags{}, fmt.Errorf("decode ID3 tag size: %w", err)
	}
	if tagSize > maxTagBytes {
		return Tags{}, fmt.Errorf("ID3 tag size %d exceeds %d-byte limit", tagSize, maxTagBytes)
	}
	payload := make([]byte, tagSize)
	if _, err := io.ReadFull(source, payload); err != nil {
		return Tags{}, fmt.Errorf("read ID3 tag payload: %w", err)
	}
	if header[5]&0x80 != 0 {
		payload = removeUnsynchronization(payload)
	}
	if header[5]&0x40 != 0 {
		payload, err = stripID3ExtendedHeader(version, payload)
		if err != nil {
			return Tags{}, err
		}
	}

	tags := Tags{TagFormat: fmt.Sprintf("id3v2.%d", version)}
	for len(payload) >= 10 {
		frameID := string(payload[:4])
		if frameID == "\x00\x00\x00\x00" {
			break
		}
		if !validID3FrameID(frameID) {
			break
		}

		var frameSize int
		if version == 4 {
			size, err := decodeSynchsafe32(payload[4:8])
			if err != nil {
				return Tags{}, fmt.Errorf("decode ID3 frame %s size: %w", frameID, err)
			}
			frameSize = size
		} else {
			frameSize = int(binary.BigEndian.Uint32(payload[4:8]))
		}
		if frameSize < 0 || frameSize > maxTagBytes || frameSize > len(payload)-10 {
			return Tags{}, fmt.Errorf("ID3 frame %s has invalid size %d", frameID, frameSize)
		}
		frameFlags := payload[8:10]
		frameData := payload[10 : 10+frameSize]
		payload = payload[10+frameSize:]
		if frameSize == 0 {
			continue
		}

		if version == 3 && (frameFlags[1]&0xC0) != 0 { // compression or encryption
			continue
		}
		if version == 4 {
			if (frameFlags[1] & 0x0C) != 0 { // compression or encryption
				continue
			}
			if frameFlags[1]&0x02 != 0 {
				frameData = removeUnsynchronization(frameData)
			}
			if frameFlags[1]&0x01 != 0 { // data-length indicator
				if len(frameData) < 4 {
					return Tags{}, fmt.Errorf("ID3 frame %s has truncated data-length indicator", frameID)
				}
				frameData = frameData[4:]
			}
		}

		if !isMetadataTextFrame(frameID) {
			continue
		}
		value, err := decodeID3Text(frameData)
		if err != nil {
			return Tags{}, fmt.Errorf("decode ID3 frame %s: %w", frameID, err)
		}
		assignID3Text(&tags, frameID, value)
	}
	return tags, nil
}

func stripID3ExtendedHeader(version byte, payload []byte) ([]byte, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("truncated ID3 extended header")
	}
	if version == 3 {
		size := int(binary.BigEndian.Uint32(payload[:4]))
		total := 4 + size
		if size < 6 || total > len(payload) {
			return nil, fmt.Errorf("invalid ID3v2.3 extended header size %d", size)
		}
		return payload[total:], nil
	}
	size, err := decodeSynchsafe32(payload[:4])
	if err != nil {
		return nil, fmt.Errorf("decode ID3v2.4 extended header size: %w", err)
	}
	if size < 6 || size > len(payload) {
		return nil, fmt.Errorf("invalid ID3v2.4 extended header size %d", size)
	}
	return payload[size:], nil
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

func removeUnsynchronization(value []byte) []byte {
	out := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		out = append(out, value[i])
		if value[i] == 0xFF && i+1 < len(value) && value[i+1] == 0x00 {
			i++
		}
	}
	return out
}

func validID3FrameID(value string) bool {
	if len(value) != 4 {
		return false
	}
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func isMetadataTextFrame(frameID string) bool {
	switch frameID {
	case "TIT2", "TPE1", "TALB", "TPE2", "TCON", "TRCK", "TPOS", "TDRC", "TYER":
		return true
	default:
		return false
	}
}

func assignID3Text(tags *Tags, frameID, value string) {
	value = normalizeText(value)
	if value == "" {
		return
	}
	switch frameID {
	case "TIT2":
		tags.Title = firstValue(tags.Title, value)
	case "TPE1":
		tags.Artist = firstValue(tags.Artist, value)
	case "TALB":
		tags.Album = firstValue(tags.Album, value)
	case "TPE2":
		tags.AlbumArtist = firstValue(tags.AlbumArtist, value)
	case "TCON":
		tags.Genre = firstValue(tags.Genre, value)
	case "TRCK":
		tags.TrackNumber = firstValue(tags.TrackNumber, value)
	case "TPOS":
		tags.DiscNumber = firstValue(tags.DiscNumber, value)
	case "TDRC", "TYER":
		tags.Date = firstValue(tags.Date, value)
	}
}

func decodeID3Text(frame []byte) (string, error) {
	if len(frame) < 1 {
		return "", nil
	}
	data := frame[1:]
	if len(data) > maxTextBytes {
		return "", fmt.Errorf("text frame exceeds %d-byte limit", maxTextBytes)
	}
	switch frame[0] {
	case 0:
		runes := make([]rune, len(data))
		for i, b := range data {
			runes[i] = rune(b)
		}
		return string(runes), nil
	case 1:
		if len(data) < 2 {
			return "", fmt.Errorf("UTF-16 text frame is missing BOM")
		}
		var order binary.ByteOrder
		switch {
		case data[0] == 0xFF && data[1] == 0xFE:
			order = binary.LittleEndian
		case data[0] == 0xFE && data[1] == 0xFF:
			order = binary.BigEndian
		default:
			return "", fmt.Errorf("UTF-16 text frame has no BOM")
		}
		return decodeUTF16(data[2:], order)
	case 2:
		return decodeUTF16(data, binary.BigEndian)
	case 3:
		if !utf8.Valid(data) {
			return "", fmt.Errorf("invalid UTF-8 text frame")
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported ID3 text encoding %d", frame[0])
	}
}

func decodeUTF16(data []byte, order binary.ByteOrder) (string, error) {
	if len(data)%2 != 0 {
		return "", fmt.Errorf("UTF-16 text frame has odd byte length")
	}
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = order.Uint16(data[i*2 : i*2+2])
	}
	return string(utf16.Decode(units)), nil
}

func normalizeText(value string) string {
	if index := strings.IndexRune(value, '\x00'); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}

func firstValue(existing, candidate string) string {
	if existing != "" {
		return existing
	}
	return candidate
}

func extractFLAC(source io.Reader) (Tags, error) {
	magic := make([]byte, 4)
	if _, err := io.ReadFull(source, magic); err != nil {
		return Tags{}, fmt.Errorf("read FLAC signature: %w", err)
	}
	if string(magic) != "fLaC" {
		return Tags{}, fmt.Errorf("source does not contain a FLAC signature")
	}

	for blockIndex := 0; blockIndex < 128; blockIndex++ {
		header := make([]byte, 4)
		if _, err := io.ReadFull(source, header); err != nil {
			return Tags{}, fmt.Errorf("read FLAC metadata block header: %w", err)
		}
		last := header[0]&0x80 != 0
		blockType := header[0] & 0x7F
		length := int(header[1])<<16 | int(header[2])<<8 | int(header[3])
		if length > maxTagBytes {
			return Tags{}, fmt.Errorf("FLAC metadata block size %d exceeds %d-byte limit", length, maxTagBytes)
		}
		if blockType == 4 {
			payload := make([]byte, length)
			if _, err := io.ReadFull(source, payload); err != nil {
				return Tags{}, fmt.Errorf("read FLAC Vorbis comment block: %w", err)
			}
			return parseVorbisComments(payload)
		}
		if _, err := io.CopyN(io.Discard, source, int64(length)); err != nil {
			return Tags{}, fmt.Errorf("skip FLAC metadata block: %w", err)
		}
		if last {
			return Tags{TagFormat: "none"}, nil
		}
	}
	return Tags{}, fmt.Errorf("FLAC metadata block count exceeds limit")
}

func parseVorbisComments(payload []byte) (Tags, error) {
	cursor := 0
	readUint32 := func(label string) (uint32, error) {
		if cursor+4 > len(payload) {
			return 0, fmt.Errorf("truncated FLAC Vorbis %s", label)
		}
		value := binary.LittleEndian.Uint32(payload[cursor : cursor+4])
		cursor += 4
		return value, nil
	}
	readBytes := func(length uint32, label string) ([]byte, error) {
		if length > maxTagBytes || uint64(cursor)+uint64(length) > uint64(len(payload)) {
			return nil, fmt.Errorf("invalid FLAC Vorbis %s length %d", label, length)
		}
		value := payload[cursor : cursor+int(length)]
		cursor += int(length)
		return value, nil
	}

	vendorLength, err := readUint32("vendor length")
	if err != nil {
		return Tags{}, err
	}
	vendor, err := readBytes(vendorLength, "vendor")
	if err != nil {
		return Tags{}, err
	}
	if !utf8.Valid(vendor) {
		return Tags{}, fmt.Errorf("FLAC Vorbis vendor is not valid UTF-8")
	}
	count, err := readUint32("comment count")
	if err != nil {
		return Tags{}, err
	}
	if count > maxVorbisFields {
		return Tags{}, fmt.Errorf("FLAC Vorbis comment count %d exceeds limit", count)
	}

	tags := Tags{TagFormat: "flac-vorbis-comment"}
	for i := uint32(0); i < count; i++ {
		length, err := readUint32("comment length")
		if err != nil {
			return Tags{}, err
		}
		if length > maxTextBytes {
			return Tags{}, fmt.Errorf("FLAC Vorbis comment %d exceeds %d-byte text limit", i, maxTextBytes)
		}
		raw, err := readBytes(length, "comment")
		if err != nil {
			return Tags{}, err
		}
		if !utf8.Valid(raw) {
			return Tags{}, fmt.Errorf("FLAC Vorbis comment %d is not valid UTF-8", i)
		}
		key, value, ok := strings.Cut(string(raw), "=")
		if !ok {
			continue
		}
		assignVorbisField(&tags, strings.ToUpper(strings.TrimSpace(key)), normalizeText(value))
	}
	return tags, nil
}

func assignVorbisField(tags *Tags, key, value string) {
	if value == "" {
		return
	}
	switch key {
	case "TITLE":
		tags.Title = firstValue(tags.Title, value)
	case "ARTIST":
		tags.Artist = firstValue(tags.Artist, value)
	case "ALBUM":
		tags.Album = firstValue(tags.Album, value)
	case "ALBUMARTIST", "ALBUM ARTIST":
		tags.AlbumArtist = firstValue(tags.AlbumArtist, value)
	case "GENRE":
		tags.Genre = firstValue(tags.Genre, value)
	case "DATE", "YEAR":
		tags.Date = firstValue(tags.Date, value)
	case "TRACKNUMBER":
		tags.TrackNumber = firstValue(tags.TrackNumber, value)
	case "DISCNUMBER":
		tags.DiscNumber = firstValue(tags.DiscNumber, value)
	}
}
