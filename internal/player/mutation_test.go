package player

import "testing"

func TestQueueAppendStartsEmptyQueue(t *testing.T) {
	q := NewQueue(nil)
	if err := q.Append("one", "two"); err != nil {
		t.Fatal(err)
	}
	if q.Current != 0 {
		t.Fatalf("current = %d, want 0", q.Current)
	}
	if id, ok := q.CurrentTrackID(); !ok || id != "one" {
		t.Fatalf("current track = %q ok=%v, want one", id, ok)
	}
}

func TestQueueAppendRejectsBlankTrackWithoutMutation(t *testing.T) {
	q := NewQueue([]string{"one"})
	if err := q.Append("two", "  "); err != ErrInvalidTrackID {
		t.Fatalf("error = %v, want %v", err, ErrInvalidTrackID)
	}
	if len(q.TrackIDs) != 1 || q.TrackIDs[0] != "one" {
		t.Fatalf("queue mutated after rejected append: %v", q.TrackIDs)
	}
}

func TestQueueRemoveBeforeCurrentPreservesCurrentTrack(t *testing.T) {
	q := NewQueue([]string{"one", "two", "three"})
	if err := q.SetCurrent(2); err != nil {
		t.Fatal(err)
	}
	removed, err := q.Remove(0)
	if err != nil {
		t.Fatal(err)
	}
	if removed != "one" || q.Current != 1 {
		t.Fatalf("removed=%q current=%d, want one and 1", removed, q.Current)
	}
	if id, _ := q.CurrentTrackID(); id != "three" {
		t.Fatalf("current track = %q, want three", id)
	}
}

func TestQueueRemoveCurrentSelectsNextWhenAvailable(t *testing.T) {
	q := NewQueue([]string{"one", "two", "three"})
	if err := q.SetCurrent(1); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Remove(1); err != nil {
		t.Fatal(err)
	}
	if id, _ := q.CurrentTrackID(); id != "three" {
		t.Fatalf("current track = %q, want three", id)
	}
}

func TestQueueRemoveCurrentTailFallsBackToPrevious(t *testing.T) {
	q := NewQueue([]string{"one", "two"})
	if err := q.SetCurrent(1); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Remove(1); err != nil {
		t.Fatal(err)
	}
	if q.Current != 0 {
		t.Fatalf("current = %d, want 0", q.Current)
	}
	if id, _ := q.CurrentTrackID(); id != "one" {
		t.Fatalf("current track = %q, want one", id)
	}
}

func TestQueueRemoveLastTrackEmptiesQueue(t *testing.T) {
	q := NewQueue([]string{"one"})
	if _, err := q.Remove(0); err != nil {
		t.Fatal(err)
	}
	if len(q.TrackIDs) != 0 || q.Current != -1 {
		t.Fatalf("queue = %v current=%d, want empty and -1", q.TrackIDs, q.Current)
	}
}
