package preferences

import (
	"reflect"
	"testing"
)

func TestFavoriteTracksAddRemoveAndSnapshot(t *testing.T) {
	favorites, err := NewFavoriteTracks("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if added, err := favorites.Add("track-b"); err != nil || !added {
		t.Fatalf("add result=%v err=%v", added, err)
	}
	if added, err := favorites.Add("track-a"); err != nil || !added {
		t.Fatalf("add result=%v err=%v", added, err)
	}
	if added, err := favorites.Add("track-a"); err != nil || added {
		t.Fatalf("duplicate add result=%v err=%v", added, err)
	}
	if !favorites.Contains("track-a") || favorites.Contains(" track-a ") {
		t.Fatal("unexpected contains result")
	}

	snapshot, err := favorites.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snapshot.TrackIDs, []string{"track-a", "track-b"}) {
		t.Fatalf("snapshot=%+v", snapshot)
	}
	snapshot.TrackIDs[0] = "changed"
	if !favorites.Contains("track-a") {
		t.Fatal("snapshot mutated favorite set")
	}

	snapshot, _ = favorites.Snapshot()
	restored, err := RestoreFavoriteTracks(snapshot)
	if err != nil || restored.Count() != 2 {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
	removed, err := restored.Remove("track-a")
	if err != nil || !removed || restored.Contains("track-a") {
		t.Fatalf("remove result=%v err=%v", removed, err)
	}
}

func TestFavoriteTracksRejectsMalformedState(t *testing.T) {
	if _, err := NewFavoriteTracks(" "); err == nil {
		t.Fatal("expected blank user scope rejection")
	}
	if _, err := RestoreFavoriteTracks(FavoriteTracksSnapshot{
		UserID:   "user-1",
		TrackIDs: []string{"track-a", "track-a"},
	}); err == nil {
		t.Fatal("expected duplicate snapshot rejection")
	}
	if _, err := RestoreFavoriteTracks(FavoriteTracksSnapshot{
		UserID:   "user-1",
		TrackIDs: []string{" track-a "},
	}); err == nil {
		t.Fatal("expected whitespace-altered track rejection")
	}

	var zero FavoriteTracks
	if _, err := zero.Add("track-a"); err == nil {
		t.Fatal("expected zero-value state rejection")
	}
	if _, err := zero.Snapshot(); err == nil {
		t.Fatal("expected zero-value snapshot rejection")
	}
	var nilFavorites *FavoriteTracks
	if _, err := nilFavorites.Add("track-a"); err == nil {
		t.Fatal("expected nil receiver rejection")
	}
}
