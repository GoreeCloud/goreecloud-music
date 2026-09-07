package player

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRestoreQueueRecordRoundTrip(t *testing.T) {
	q := Queue{TrackIDs: []string{"a", "b"}, Current: 1, Repeat: RepeatAll}
	r, err := NewQueueRecord(q)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreQueueRecord(7, r.Payload())
	if err != nil {
		t.Fatal(err)
	}
	if restored.Revision() != 7 {
		t.Fatalf("revision=%d", restored.Revision())
	}
	got, err := restored.Restore()
	if err != nil || !reflect.DeepEqual(got, q) {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestRestoreQueueRecordRejectsNonCanonicalAndOversized(t *testing.T) {
	noncanonical := []byte(`{ "schemaVersion":1,"trackIds":["a"],"current":0,"repeat":"off"}`)
	if _, err := RestoreQueueRecord(0, noncanonical); !errors.Is(err, ErrInvalidQueueRecord) {
		t.Fatalf("noncanonical err=%v", err)
	}
	if _, err := RestoreQueueRecord(0, []byte(strings.Repeat("x", MaxQueueRecordPayloadBytes+1))); !errors.Is(err, ErrInvalidQueueRecord) {
		t.Fatalf("oversize err=%v", err)
	}
}

func TestNewQueueRecordEnforcesPayloadBound(t *testing.T) {
	tracks := make([]string, 1000)
	for i := range tracks {
		tracks[i] = strings.Repeat("x", 80)
	}
	if _, err := NewQueueRecord(Queue{TrackIDs: tracks, Current: 0, Repeat: RepeatOff}); !errors.Is(err, ErrInvalidQueueRecord) {
		t.Fatalf("err=%v", err)
	}
}
