package player

import "testing"

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
