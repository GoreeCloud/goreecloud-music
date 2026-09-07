package playlists

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestPlaylistPersistenceRoundTripPreservesOrderAndDuplicates(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	for _, trackID := range []string{"a", "b", "a"} {
		_ = playlist.Append(trackID)
	}
	payload, err := Encode(playlist)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"schemaVersion":1,"userId":"user-1","id":"playlist-1","name":"Mix","trackIds":["a","b","a"]}`
	if string(payload) != want {
		t.Fatalf("payload=%s", payload)
	}
	restored, err := Decode(payload)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, _ := restored.Snapshot()
	if !reflect.DeepEqual(snapshot.TrackIDs, []string{"a", "b", "a"}) {
		t.Fatalf("tracks=%v", snapshot.TrackIDs)
	}
}

func TestPlaylistPersistenceRejectsUnknownTrailingAndNoncanonicalJSON(t *testing.T) {
	cases := []string{
		`{"schemaVersion":1,"userId":"user","id":"p","name":"Mix","trackIds":[],"extra":true}`,
		`{"schemaVersion":1,"userId":"user","id":"p","name":"Mix","trackIds":[]} {}`,
		`{ "schemaVersion":1,"userId":"user","id":"p","name":"Mix","trackIds":[]}`,
		`{"schemaVersion":1,"userId":"user","id":"p","name":"Mix","trackIds":null}`,
	}
	for i, payload := range cases {
		if _, err := Decode([]byte(payload)); !errors.Is(err, ErrInvalidPlaylistPayload) {
			t.Fatalf("case=%d error=%v", i, err)
		}
	}
}

func TestPlaylistPersistenceRejectsSchemaAndMalformedState(t *testing.T) {
	cases := []string{
		`{"schemaVersion":2,"userId":"user","id":"p","name":"Mix","trackIds":[]}`,
		`{"schemaVersion":1,"userId":" user","id":"p","name":"Mix","trackIds":[]}`,
		`{"schemaVersion":1,"userId":"user","id":"p","name":" Mix","trackIds":[]}`,
		`{"schemaVersion":1,"userId":"user","id":"p","name":"Mix","trackIds":[" bad"]}`,
	}
	for i, payload := range cases {
		if _, err := Decode([]byte(payload)); !errors.Is(err, ErrInvalidPlaylistPayload) {
			t.Fatalf("case=%d error=%v", i, err)
		}
	}
}

func TestPlaylistPersistenceEnforcesPayloadBound(t *testing.T) {
	if _, err := Decode(make([]byte, MaxPlaylistPayloadBytes+1)); !errors.Is(err, ErrInvalidPlaylistPayload) {
		t.Fatalf("error=%v", err)
	}

	playlist, _ := New("user", "playlist", "Mix")
	largeTrackID := strings.Repeat("x", MaxPlaylistPayloadBytes)
	_ = playlist.Append(largeTrackID)
	if _, err := Encode(playlist); !errors.Is(err, ErrInvalidPlaylistPayload) {
		t.Fatalf("error=%v", err)
	}
}
