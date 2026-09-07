package player

import (
	"errors"
	"reflect"
	"testing"
)

func TestQueueRecordRoundTripStartsAtRevisionZero(t *testing.T) {
	queue := Queue{TrackIDs: []string{"track-a", "track-b"}, Current: 1, Repeat: RepeatAll}
	record, err := NewQueueRecord(queue)
	if err != nil {
		t.Fatal(err)
	}
	if record.Revision() != 0 {
		t.Fatalf("revision = %d, want 0", record.Revision())
	}
	restored, err := record.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, queue) {
		t.Fatalf("restored queue = %+v, want %+v", restored, queue)
	}
}

func TestQueueRecordUpdateAdvancesRevision(t *testing.T) {
	record, err := NewQueueRecord(NewQueue([]string{"track-a"}))
	if err != nil {
		t.Fatal(err)
	}
	updatedQueue := Queue{TrackIDs: []string{"track-a", "track-b"}, Current: 1, Repeat: RepeatOne}
	updated, err := record.Update(0, updatedQueue)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Revision() != 1 {
		t.Fatalf("revision = %d, want 1", updated.Revision())
	}
	restored, err := updated.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, updatedQueue) {
		t.Fatalf("restored queue = %+v, want %+v", restored, updatedQueue)
	}
}

func TestQueueRecordStaleUpdateLeavesRecordUnchanged(t *testing.T) {
	record, err := NewQueueRecord(NewQueue([]string{"track-a"}))
	if err != nil {
		t.Fatal(err)
	}
	before := record.Payload()

	updated, err := record.Update(1, NewQueue([]string{"track-b"}))
	if !errors.Is(err, ErrStaleQueueRevision) {
		t.Fatalf("error = %v, want stale revision", err)
	}
	if updated.Revision() != record.Revision() || !reflect.DeepEqual(updated.Payload(), before) {
		t.Fatal("stale update changed the record")
	}
}

func TestQueueRecordInvalidQueueLeavesRecordUnchanged(t *testing.T) {
	record, err := NewQueueRecord(NewQueue([]string{"track-a"}))
	if err != nil {
		t.Fatal(err)
	}
	before := record.Payload()
	invalid := Queue{TrackIDs: []string{"track-a", " "}, Current: 0, Repeat: RepeatOff}

	updated, err := record.Update(0, invalid)
	if !errors.Is(err, ErrInvalidQueueSnapshot) {
		t.Fatalf("error = %v, want invalid snapshot", err)
	}
	if updated.Revision() != record.Revision() || !reflect.DeepEqual(updated.Payload(), before) {
		t.Fatal("invalid update changed the record")
	}
}

func TestQueueRecordPayloadIsDefensivelyCopied(t *testing.T) {
	record, err := NewQueueRecord(NewQueue([]string{"track-a"}))
	if err != nil {
		t.Fatal(err)
	}
	payload := record.Payload()
	payload[0] ^= 0xff

	if _, err := record.Restore(); err != nil {
		t.Fatalf("mutating payload copy corrupted record: %v", err)
	}
}
