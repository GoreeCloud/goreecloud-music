package playlists

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestRecordUpdateAdvancesRevisionAndPreservesPlaylistState(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	_ = playlist.Append("a")
	record, err := NewRecord(playlist)
	if err != nil || record.Revision() != 0 {
		t.Fatalf("record=%+v err=%v", record, err)
	}

	_ = playlist.Rename("Updated")
	_ = playlist.Append("b")
	next, err := record.Update(0, playlist)
	if err != nil || next.Revision() != 1 {
		t.Fatalf("next=%+v err=%v", next, err)
	}
	restored, err := next.Restore()
	if err != nil || restored.Name() != "Updated" || !reflect.DeepEqual(restored.TrackIDs(), []string{"a", "b"}) {
		t.Fatalf("name=%q tracks=%v err=%v", restored.Name(), restored.TrackIDs(), err)
	}
}

func TestRecordRejectsStaleWriterAndIdentityReplacement(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	if _, err := record.Update(1, playlist); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("error=%v", err)
	}

	otherUser, _ := New("user-2", "playlist-1", "Mix")
	if _, err := record.Update(0, otherUser); !errors.Is(err, ErrInvalidRecord) {
		t.Fatalf("user error=%v", err)
	}
	otherID, _ := New("user-1", "playlist-2", "Mix")
	if _, err := record.Update(0, otherID); !errors.Is(err, ErrInvalidRecord) {
		t.Fatalf("id error=%v", err)
	}
}

func TestRecordCopiesPayloadAndRejectsOverflow(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	payload, err := Encode(playlist)
	if err != nil {
		t.Fatal(err)
	}
	record, err := RestoreRecord(7, payload)
	if err != nil || record.Revision() != 7 {
		t.Fatalf("record=%+v err=%v", record, err)
	}
	copyOut := record.Payload()
	copyOut[0] = 'x'
	payload[0] = 'x'
	if _, err := record.Restore(); err != nil {
		t.Fatal("payload mutation escaped into record")
	}

	record.revision = Revision(math.MaxUint64)
	if _, err := record.Update(record.Revision(), playlist); !errors.Is(err, ErrRevisionExhausted) {
		t.Fatalf("error=%v", err)
	}
}

func TestRestoreRecordRejectsMalformedPayload(t *testing.T) {
	if _, err := RestoreRecord(1, nil); !errors.Is(err, ErrInvalidRecord) {
		t.Fatalf("error=%v", err)
	}
	if _, err := RestoreRecord(1, []byte(`{"schemaVersion":1}`)); !errors.Is(err, ErrInvalidRecord) {
		t.Fatalf("error=%v", err)
	}
}
