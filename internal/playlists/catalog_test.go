package playlists

import (
	"errors"
	"testing"
)

func TestCatalogAddsListsAndReturnsDefensivePlaylists(t *testing.T) {
	catalog, err := NewCatalog("user-1")
	if err != nil {
		t.Fatal(err)
	}
	first := mustPlaylist(t, "user-1", "playlist-b", "Beta", "track-2")
	second := mustPlaylist(t, "user-1", "playlist-a", "Alpha", "track-1", "track-1")
	if err := catalog.Add(first); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Add(second); err != nil {
		t.Fatal(err)
	}

	entries := catalog.Entries()
	if len(entries) != 2 || entries[0].ID != "playlist-a" || entries[0].TrackCount != 2 || entries[1].ID != "playlist-b" {
		t.Fatalf("entries=%+v", entries)
	}

	clone, ok := catalog.Get("playlist-a")
	if !ok {
		t.Fatal("expected playlist")
	}
	if err := clone.Rename("Changed"); err != nil {
		t.Fatal(err)
	}
	if err := clone.Append("track-new"); err != nil {
		t.Fatal(err)
	}
	original, ok := catalog.Get("playlist-a")
	if !ok || original.Name() != "Alpha" || original.Len() != 2 {
		t.Fatalf("original=%+v ok=%v", original, ok)
	}
}

func TestCatalogSnapshotRestoreIsDeterministicAndDefensive(t *testing.T) {
	catalog, err := NewCatalog("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.Add(mustPlaylist(t, "user-1", "playlist-b", "Beta", "track-2")); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Add(mustPlaylist(t, "user-1", "playlist-a", "Alpha", "track-1", "track-1")); err != nil {
		t.Fatal(err)
	}

	snapshot, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Playlists) != 2 || snapshot.Playlists[0].ID != "playlist-a" || snapshot.Playlists[1].ID != "playlist-b" {
		t.Fatalf("snapshot=%+v", snapshot)
	}

	snapshot.Playlists[0].TrackIDs[0] = "external-mutation"
	original, ok := catalog.Get("playlist-a")
	if !ok || original.TrackIDs()[0] != "track-1" {
		t.Fatalf("original=%+v ok=%v", original, ok)
	}

	snapshot, err = catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreCatalog(snapshot)
	if err != nil || restored.Count() != 2 {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
	restoredPlaylist, ok := restored.Get("playlist-a")
	if !ok || restoredPlaylist.Len() != 2 || restoredPlaylist.TrackIDs()[0] != "track-1" || restoredPlaylist.TrackIDs()[1] != "track-1" {
		t.Fatalf("playlist=%+v ok=%v", restoredPlaylist, ok)
	}
}

func TestCatalogRejectsDuplicateCrossUserAndMalformedState(t *testing.T) {
	catalog, err := NewCatalog("user-1")
	if err != nil {
		t.Fatal(err)
	}
	playlist := mustPlaylist(t, "user-1", "playlist-1", "One")
	if err := catalog.Add(playlist); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Add(playlist); !errors.Is(err, ErrInvalidPlaylistCatalog) {
		t.Fatalf("duplicate error=%v", err)
	}
	if err := catalog.Add(mustPlaylist(t, "user-2", "playlist-2", "Two")); !errors.Is(err, ErrInvalidPlaylistCatalog) {
		t.Fatalf("cross-user error=%v", err)
	}

	bad := CatalogSnapshot{
		UserID: "user-1",
		Playlists: []Snapshot{
			{UserID: "user-1", ID: "playlist-1", Name: "One", TrackIDs: []string{}},
			{UserID: "user-1", ID: "playlist-1", Name: "Duplicate", TrackIDs: []string{}},
		},
	}
	if _, err := RestoreCatalog(bad); !errors.Is(err, ErrInvalidPlaylistCatalog) {
		t.Fatalf("restore error=%v", err)
	}

	var zero Catalog
	if err := zero.Add(playlist); !errors.Is(err, ErrInvalidPlaylistCatalog) {
		t.Fatalf("zero add error=%v", err)
	}
	if zero.Entries() != nil {
		t.Fatal("zero catalog should not expose entries")
	}
}

func TestCatalogRemoveIsIdempotentForAbsentPlaylist(t *testing.T) {
	catalog, err := NewCatalog("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.Add(mustPlaylist(t, "user-1", "playlist-1", "One")); err != nil {
		t.Fatal(err)
	}
	removed, err := catalog.Remove("playlist-1")
	if err != nil || !removed || catalog.Count() != 0 {
		t.Fatalf("removed=%v count=%d err=%v", removed, catalog.Count(), err)
	}
	removed, err = catalog.Remove("playlist-1")
	if err != nil || removed {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
}

func mustPlaylist(t *testing.T, userID, id, name string, trackIDs ...string) Playlist {
	t.Helper()
	playlist, err := New(userID, id, name)
	if err != nil {
		t.Fatal(err)
	}
	for _, trackID := range trackIDs {
		if err := playlist.Append(trackID); err != nil {
			t.Fatal(err)
		}
	}
	return playlist
}
