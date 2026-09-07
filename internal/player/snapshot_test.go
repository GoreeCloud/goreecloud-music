package player

import (
	"errors"
	"reflect"
	"testing"
)

func TestQueueSnapshotRoundTrip(t *testing.T) {
	q := NewQueue([]string{"one", "two", "three"})
	q.Current = 1
	q.Repeat = RepeatAll
	s := q.Snapshot()
	restored, err := RestoreQueue(s)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored.TrackIDs, q.TrackIDs) || restored.Current != 1 || restored.Repeat != RepeatAll {
		t.Fatalf("restored=%+v want=%+v", restored, q)
	}
}

func TestQueueSnapshotCopiesTrackIDs(t *testing.T) {
	q := NewQueue([]string{"one", "two"})
	s := q.Snapshot()
	s.TrackIDs[0] = "changed"
	if q.TrackIDs[0] != "one" {
		t.Fatalf("snapshot mutated live queue: %v", q.TrackIDs)
	}
	restored, err := RestoreQueue(q.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	restored.TrackIDs[0] = "changed"
	if q.TrackIDs[0] != "one" {
		t.Fatalf("restored queue shares live storage: %v", q.TrackIDs)
	}
}

func TestRestoreQueueRejectsInvalidSnapshot(t *testing.T) {
	tests := []QueueSnapshot{
		{TrackIDs: nil, Current: 0, Repeat: RepeatOff},
		{TrackIDs: []string{"one"}, Current: -1, Repeat: RepeatOff},
		{TrackIDs: []string{"one"}, Current: 1, Repeat: RepeatOff},
		{TrackIDs: []string{"one"}, Current: 0, Repeat: RepeatMode("future")},
	}
	for _, s := range tests {
		if _, err := RestoreQueue(s); !errors.Is(err, ErrInvalidQueueSnapshot) {
			t.Fatalf("snapshot=%+v err=%v", s, err)
		}
	}
}

func TestRestoreQueueAcceptsEmptyQueue(t *testing.T) {
	q, err := RestoreQueue(QueueSnapshot{Current: -1, Repeat: RepeatOff})
	if err != nil {
		t.Fatal(err)
	}
	if len(q.TrackIDs) != 0 || q.Current != -1 {
		t.Fatalf("queue=%+v", q)
	}
}
