package player

import (
	"errors"
	"reflect"
	"testing"
)

func TestQueueSnapshotJSONRoundTrip(t *testing.T) {
	queue := Queue{TrackIDs: []string{"track-a", "track-b"}, Current: 1, Repeat: RepeatAll}
	data, err := EncodeQueueSnapshot(queue)
	if err != nil {
		t.Fatal(err)
	}

	restored, err := DecodeQueueSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, queue) {
		t.Fatalf("restored queue = %+v, want %+v", restored, queue)
	}
}

func TestQueueSnapshotDecodeRejectsUnsupportedVersion(t *testing.T) {
	_, err := DecodeQueueSnapshot([]byte(`{"schemaVersion":2,"trackIds":["track-a"],"current":0,"repeat":"off"}`))
	if !errors.Is(err, ErrUnsupportedQueueSnapshotVersion) {
		t.Fatalf("error = %v, want unsupported version", err)
	}
}

func TestQueueSnapshotDecodeRejectsUnknownFieldsAndTrailingValues(t *testing.T) {
	cases := [][]byte{
		[]byte(`{"schemaVersion":1,"trackIds":["track-a"],"current":0,"repeat":"off","extra":true}`),
		[]byte(`{"schemaVersion":1,"trackIds":["track-a"],"current":0,"repeat":"off"} {}`),
	}
	for _, data := range cases {
		if _, err := DecodeQueueSnapshot(data); !errors.Is(err, ErrInvalidQueueSnapshot) {
			t.Fatalf("payload %s error = %v, want invalid snapshot", data, err)
		}
	}
}

func TestQueueSnapshotRejectsBlankTrackIDs(t *testing.T) {
	snapshot := QueueSnapshot{TrackIDs: []string{"track-a", " "}, Current: 0, Repeat: RepeatOff}
	if _, err := RestoreQueue(snapshot); !errors.Is(err, ErrInvalidQueueSnapshot) {
		t.Fatalf("restore error = %v, want invalid snapshot", err)
	}
	if _, err := EncodeQueueSnapshot(Queue{TrackIDs: snapshot.TrackIDs, Current: 0, Repeat: RepeatOff}); !errors.Is(err, ErrInvalidQueueSnapshot) {
		t.Fatalf("encode error = %v, want invalid snapshot", err)
	}
}

func TestQueueSnapshotDecodeRejectsInvalidState(t *testing.T) {
	_, err := DecodeQueueSnapshot([]byte(`{"schemaVersion":1,"trackIds":[],"current":0,"repeat":"off"}`))
	if !errors.Is(err, ErrInvalidQueueSnapshot) {
		t.Fatalf("error = %v, want invalid snapshot", err)
	}
}
