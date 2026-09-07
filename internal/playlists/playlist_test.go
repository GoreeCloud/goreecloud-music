package playlists

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestPlaylistPreservesOrderAndDuplicateTracks(t *testing.T) {
	playlist, err := New("user-1", "playlist-1", "Morning")
	if err != nil {
		t.Fatal(err)
	}
	for _, trackID := range []string{"a", "b", "a"} {
		if err := playlist.Append(trackID); err != nil {
			t.Fatal(err)
		}
	}
	if err := playlist.Insert(1, "c"); err != nil {
		t.Fatal(err)
	}
	if got := playlist.TrackIDs(); !reflect.DeepEqual(got, []string{"a", "c", "b", "a"}) {
		t.Fatalf("tracks=%v", got)
	}
	if err := playlist.Move(3, 1); err != nil {
		t.Fatal(err)
	}
	if got := playlist.TrackIDs(); !reflect.DeepEqual(got, []string{"a", "a", "c", "b"}) {
		t.Fatalf("tracks=%v", got)
	}
	removed, err := playlist.RemoveAt(2)
	if err != nil || removed != "c" {
		t.Fatalf("removed=%q err=%v", removed, err)
	}
}

func TestPlaylistRenameAndSnapshotAreDefensive(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Morning")
	_ = playlist.Append("a")
	if err := playlist.Rename("Evening"); err != nil {
		t.Fatal(err)
	}
	snapshot, err := playlist.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Name != "Evening" || !reflect.DeepEqual(snapshot.TrackIDs, []string{"a"}) {
		t.Fatalf("snapshot=%+v", snapshot)
	}
	snapshot.TrackIDs[0] = "mutated"
	if got := playlist.TrackIDs(); !reflect.DeepEqual(got, []string{"a"}) {
		t.Fatalf("tracks=%v", got)
	}
}

func TestRestorePlaylistPreservesExactSnapshot(t *testing.T) {
	snapshot := Snapshot{UserID: "user-1", ID: "playlist-1", Name: "Mix", TrackIDs: []string{"a", "a", "b"}}
	playlist, err := Restore(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := playlist.Snapshot()
	if err != nil || !reflect.DeepEqual(restored, snapshot) {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
	snapshot.TrackIDs[0] = "changed"
	if playlist.TrackIDs()[0] != "a" {
		t.Fatal("restore retained caller slice")
	}
}

func TestPlaylistRejectsMalformedStateAndBounds(t *testing.T) {
	if _, err := New(" user", "playlist-1", "Mix"); !errors.Is(err, ErrInvalidPlaylistState) {
		t.Fatalf("error=%v", err)
	}
	if _, err := New("user", "playlist-1", strings.Repeat("x", MaxPlaylistNameLength+1)); !errors.Is(err, ErrInvalidPlaylistState) {
		t.Fatalf("error=%v", err)
	}
	if _, err := Restore(Snapshot{UserID: "user", ID: "playlist-1", Name: "Mix", TrackIDs: []string{" bad"}}); !errors.Is(err, ErrInvalidPlaylistState) {
		t.Fatalf("error=%v", err)
	}

	playlist, _ := New("user", "playlist-1", "Mix")
	if err := playlist.Insert(1, "a"); !errors.Is(err, ErrInvalidPlaylistState) {
		t.Fatalf("error=%v", err)
	}
	if _, err := playlist.RemoveAt(0); !errors.Is(err, ErrInvalidPlaylistState) {
		t.Fatalf("error=%v", err)
	}
	if err := playlist.Move(0, 0); !errors.Is(err, ErrInvalidPlaylistState) {
		t.Fatalf("error=%v", err)
	}
}
