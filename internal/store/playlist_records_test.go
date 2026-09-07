package store

import (
	"errors"
	"math"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/playlists"
)

type fakePlaylistRecordRow struct {
	revision string
	payload  []byte
	err      error
}

func (f fakePlaylistRecordRow) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}
	if len(dest) != 2 {
		return errors.New("unexpected destination count")
	}
	revision, ok := dest[0].(*string)
	if !ok {
		return errors.New("unexpected revision destination")
	}
	payload, ok := dest[1].(*[]byte)
	if !ok {
		return errors.New("unexpected payload destination")
	}
	*revision = f.revision
	*payload = append([]byte(nil), f.payload...)
	return nil
}

func TestPlaylistRevisionTextRoundTripIncludesUint64Maximum(t *testing.T) {
	values := []playlists.Revision{0, 1, 42, playlists.Revision(math.MaxUint64)}
	for _, value := range values {
		encoded := formatPlaylistRevision(value)
		decoded, err := parsePlaylistRevision(encoded)
		if err != nil || decoded != value {
			t.Fatalf("value=%d encoded=%q decoded=%d err=%v", value, encoded, decoded, err)
		}
	}
}

func TestParsePlaylistRevisionRejectsNonCanonicalAndOutOfRangeValues(t *testing.T) {
	for _, value := range []string{"", "+1", "01", "-1", " 1", "1 ", "18446744073709551616"} {
		if _, err := parsePlaylistRevision(value); !errors.Is(err, playlists.ErrInvalidRecord) {
			t.Fatalf("value=%q err=%v", value, err)
		}
	}
}

func TestScanPlaylistRecordRestoresCanonicalPayloadAndIdentity(t *testing.T) {
	playlist, err := playlists.New("user-1", "playlist-1", "Road Trip")
	if err != nil {
		t.Fatal(err)
	}
	if err := playlist.Append("track-1"); err != nil {
		t.Fatal(err)
	}
	payload, err := playlists.Encode(playlist)
	if err != nil {
		t.Fatal(err)
	}
	record, err := scanPlaylistRecord(fakePlaylistRecordRow{revision: "7", payload: payload})
	if err != nil || record.Revision() != 7 {
		t.Fatalf("record=%+v err=%v", record, err)
	}
	if err := validatePlaylistRecordIdentity(record, "user-1", "playlist-1"); err != nil {
		t.Fatal(err)
	}
	if err := validatePlaylistRecordIdentity(record, "user-2", "playlist-1"); !errors.Is(err, playlists.ErrInvalidRepositoryResult) {
		t.Fatalf("identity err=%v", err)
	}
}

func TestScanPlaylistRecordRejectsMalformedRevisionPayloadAndScannerFailure(t *testing.T) {
	playlist, _ := playlists.New("user-1", "playlist-1", "Mix")
	payload, _ := playlists.Encode(playlist)
	if _, err := scanPlaylistRecord(fakePlaylistRecordRow{revision: "01", payload: payload}); !errors.Is(err, playlists.ErrInvalidRecord) {
		t.Fatalf("revision err=%v", err)
	}
	if _, err := scanPlaylistRecord(fakePlaylistRecordRow{revision: "1", payload: []byte("not-json")}); !errors.Is(err, playlists.ErrInvalidRecord) {
		t.Fatalf("payload err=%v", err)
	}
	want := errors.New("scan failed")
	if _, err := scanPlaylistRecord(fakePlaylistRecordRow{err: want}); !errors.Is(err, want) {
		t.Fatalf("scanner err=%v", err)
	}
}
