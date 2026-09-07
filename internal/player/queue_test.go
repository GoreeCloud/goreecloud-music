package player

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestQueueNavigation(t *testing.T) {
	q := NewQueue([]string{"one", "two", "three"})
	if id, ok := q.CurrentTrackID(); !ok || id != "one" {
		t.Fatalf("expected first track, got %q ok=%v", id, ok)
	}
	if !q.Next() {
		t.Fatal("expected next track")
	}
	if id, _ := q.CurrentTrackID(); id != "two" {
		t.Fatalf("expected second track, got %q", id)
	}
	if !q.Previous() {
		t.Fatal("expected previous track")
	}
	if id, _ := q.CurrentTrackID(); id != "one" {
		t.Fatalf("expected first track, got %q", id)
	}
}

func TestQueueRepeatAllWraps(t *testing.T) {
	q := NewQueue([]string{"one", "two"})
	q.Repeat = RepeatAll
	if err := q.SetCurrent(1); err != nil {
		t.Fatal(err)
	}
	if !q.Next() {
		t.Fatal("expected repeat-all wrap")
	}
	if id, _ := q.CurrentTrackID(); id != "one" {
		t.Fatalf("expected wrap to first track, got %q", id)
	}
}

func TestQueueRepeatOneStaysCurrent(t *testing.T) {
	q := NewQueue([]string{"one", "two"})
	q.Repeat = RepeatOne
	if !q.Next() {
		t.Fatal("expected repeat-one to keep current track playable")
	}
	if id, _ := q.CurrentTrackID(); id != "one" {
		t.Fatalf("expected current track to remain one, got %q", id)
	}
}

func TestQueueShufflePreservesCurrentAndPlayedPrefix(t *testing.T) {
	q := NewQueue([]string{"one", "two", "three", "four", "five", "six"})
	if err := q.SetCurrent(1); err != nil {
		t.Fatal(err)
	}

	q.Shuffle(rand.NewSource(7))

	if q.Current != 1 {
		t.Fatalf("current index = %d, want 1", q.Current)
	}
	if !reflect.DeepEqual(q.TrackIDs[:2], []string{"one", "two"}) {
		t.Fatalf("played prefix changed: %v", q.TrackIDs[:2])
	}
	if id, ok := q.CurrentTrackID(); !ok || id != "two" {
		t.Fatalf("current track = %q ok=%v, want two", id, ok)
	}

	want := map[string]bool{"three": true, "four": true, "five": true, "six": true}
	for _, id := range q.TrackIDs[2:] {
		if !want[id] {
			t.Fatalf("unexpected shuffled track %q in %v", id, q.TrackIDs[2:])
		}
		delete(want, id)
	}
	if len(want) != 0 {
		t.Fatalf("shuffle lost tracks: %v", want)
	}
}

func TestQueueShuffleNilSourceLeavesQueueUnchanged(t *testing.T) {
	q := NewQueue([]string{"one", "two", "three"})
	before := append([]string(nil), q.TrackIDs...)
	q.Shuffle(nil)
	if !reflect.DeepEqual(q.TrackIDs, before) {
		t.Fatalf("queue changed with nil source: got %v want %v", q.TrackIDs, before)
	}
}
