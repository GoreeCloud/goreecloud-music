package preferences

import (
	"reflect"
	"testing"
)

func TestTrackRatingsSetClearAndSnapshot(t *testing.T) {
	ratings, err := NewTrackRatings("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := ratings.Set("track-b", 4); err != nil || !changed {
		t.Fatalf("set=%v err=%v", changed, err)
	}
	if changed, err := ratings.Set("track-a", 5); err != nil || !changed {
		t.Fatalf("set=%v err=%v", changed, err)
	}
	if changed, err := ratings.Set("track-a", 5); err != nil || changed {
		t.Fatalf("same set=%v err=%v", changed, err)
	}
	if stars, ok := ratings.Rating("track-a"); !ok || stars != 5 {
		t.Fatalf("stars=%d ok=%v", stars, ok)
	}

	snapshot, err := ratings.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	want := []TrackRating{
		{TrackID: "track-a", Stars: 5},
		{TrackID: "track-b", Stars: 4},
	}
	if !reflect.DeepEqual(snapshot.Ratings, want) {
		t.Fatalf("snapshot=%+v", snapshot)
	}
	snapshot.Ratings[0].Stars = 1
	if stars, _ := ratings.Rating("track-a"); stars != 5 {
		t.Fatal("snapshot mutated ratings")
	}

	changed, err := ratings.Clear("track-a")
	if err != nil || !changed {
		t.Fatalf("clear=%v err=%v", changed, err)
	}
	if _, ok := ratings.Rating("track-a"); ok {
		t.Fatal("expected unrated track")
	}
	if changed, err := ratings.Clear("track-a"); err != nil || changed {
		t.Fatalf("absent clear=%v err=%v", changed, err)
	}
}

func TestTrackRatingsRestoreAndRejectMalformedState(t *testing.T) {
	restored, err := RestoreTrackRatings(TrackRatingsSnapshot{
		UserID:  "user-1",
		Ratings: []TrackRating{{TrackID: "track-a", Stars: 3}},
	})
	if err != nil || restored.Count() != 1 {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}

	cases := []TrackRatingsSnapshot{
		{UserID: " ", Ratings: nil},
		{UserID: "user-1", Ratings: []TrackRating{{TrackID: " track-a ", Stars: 3}}},
		{UserID: "user-1", Ratings: []TrackRating{{TrackID: "track-a", Stars: 0}}},
		{UserID: "user-1", Ratings: []TrackRating{{TrackID: "track-a", Stars: 6}}},
		{
			UserID: "user-1",
			Ratings: []TrackRating{
				{TrackID: "track-a", Stars: 3},
				{TrackID: "track-a", Stars: 4},
			},
		},
	}
	for index, snapshot := range cases {
		if _, err := RestoreTrackRatings(snapshot); err == nil {
			t.Fatalf("case %d expected error", index)
		}
	}

	var zero TrackRatings
	if _, err := zero.Set("track-a", 5); err == nil {
		t.Fatal("expected zero-value state rejection")
	}
	if _, err := zero.Snapshot(); err == nil {
		t.Fatal("expected zero-value snapshot rejection")
	}
	var nilRatings *TrackRatings
	if _, err := nilRatings.Set("track-a", 5); err == nil {
		t.Fatal("expected nil receiver rejection")
	}
}

func TestTrackRatingsRejectsOutOfRangeSet(t *testing.T) {
	ratings, err := NewTrackRatings("user-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, stars := range []int{0, 6, -1} {
		if changed, err := ratings.Set("track-a", stars); err == nil || changed {
			t.Fatalf("stars=%d changed=%v err=%v", stars, changed, err)
		}
	}
	if ratings.Count() != 0 {
		t.Fatalf("count=%d", ratings.Count())
	}
}
