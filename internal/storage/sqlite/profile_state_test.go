package sqlitestore

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func TestProfileStateAuthorizationIsolationAndRevocation(t *testing.T) {
	ctx := context.Background()
	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer backend.Close()

	alice := domain.ProfileID("profile:alice")
	bob := domain.ProfileID("profile:bob")
	aliceLibrary := domain.LibraryID("library:alice")
	bobLibrary := domain.LibraryID("library:bob")
	aliceRecording := domain.RecordingID("recording:alice-song")
	bobRecording := domain.RecordingID("recording:bob-song")

	if err := backend.CreateProfile(ctx, alice, "Alice"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateProfile(ctx, bob, "Bob"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, aliceLibrary, alice, "Alice Library", filepath.Join(t.TempDir(), "alice-media")); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, bobLibrary, bob, "Bob Library", filepath.Join(t.TempDir(), "bob-media")); err != nil {
		t.Fatal(err)
	}
	insertTestRecording(t, ctx, backend, aliceRecording, aliceLibrary, "Alice Song")
	insertTestRecording(t, ctx, backend, bobRecording, bobLibrary, "Bob Song")

	if err := backend.SetFavorite(ctx, alice, aliceRecording, true); err != nil {
		t.Fatalf("SetFavorite(alice) error = %v", err)
	}
	if err := backend.SetRating(ctx, alice, aliceRecording, 87); err != nil {
		t.Fatalf("SetRating(alice) error = %v", err)
	}
	playedAt := time.Date(2026, 9, 15, 20, 0, 0, 0, time.UTC)
	if err := backend.RecordPlay(ctx, alice, aliceRecording, playedAt, 42000); err != nil {
		t.Fatalf("RecordPlay(alice) error = %v", err)
	}

	if err := backend.SetFavorite(ctx, bob, aliceRecording, true); err == nil {
		t.Fatal("expected unauthorized Bob favorite mutation to fail")
	}
	if err := backend.SetRating(ctx, bob, aliceRecording, 50); err == nil {
		t.Fatal("expected unauthorized Bob rating mutation to fail")
	}
	if err := backend.RecordPlay(ctx, bob, aliceRecording, playedAt, 1000); err == nil {
		t.Fatal("expected unauthorized Bob history mutation to fail")
	}
	if err := backend.SetFavorite(ctx, alice, bobRecording, true); err == nil {
		t.Fatal("expected unauthorized Alice mutation against Bob recording to fail")
	}

	aliceFavorites, err := backend.FavoritesForProfile(ctx, alice, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(aliceFavorites) != 1 || aliceFavorites[0].RecordingID != aliceRecording {
		t.Fatalf("Alice favorites = %#v", aliceFavorites)
	}
	bobFavorites, err := backend.FavoritesForProfile(ctx, bob, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(bobFavorites) != 0 {
		t.Fatalf("Bob favorites = %#v, want empty", bobFavorites)
	}

	aliceRating, ok, err := backend.RatingForRecording(ctx, alice, aliceRecording)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || aliceRating.Rating != 87 {
		t.Fatalf("Alice rating = %#v, ok=%v", aliceRating, ok)
	}
	aliceHistory, err := backend.RecentPlays(ctx, alice, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(aliceHistory) != 1 || aliceHistory[0].RecordingID != aliceRecording || aliceHistory[0].PlayedMS != 42000 {
		t.Fatalf("Alice history = %#v", aliceHistory)
	}

	if err := backend.SetLibraryPermission(ctx, aliceLibrary, bob, LibraryPermissionRead); err != nil {
		t.Fatalf("grant Bob read access: %v", err)
	}
	if err := backend.SetFavorite(ctx, bob, aliceRecording, true); err != nil {
		t.Fatalf("SetFavorite(bob after grant) error = %v", err)
	}
	if err := backend.SetRating(ctx, bob, aliceRecording, 63); err != nil {
		t.Fatalf("SetRating(bob after grant) error = %v", err)
	}
	if err := backend.RecordPlay(ctx, bob, aliceRecording, playedAt.Add(time.Minute), 12000); err != nil {
		t.Fatalf("RecordPlay(bob after grant) error = %v", err)
	}

	bobRating, ok, err := backend.RatingForRecording(ctx, bob, aliceRecording)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || bobRating.Rating != 63 {
		t.Fatalf("Bob rating = %#v, ok=%v", bobRating, ok)
	}
	aliceRating, ok, err = backend.RatingForRecording(ctx, alice, aliceRecording)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || aliceRating.Rating != 87 {
		t.Fatalf("Alice rating changed by Bob state: %#v, ok=%v", aliceRating, ok)
	}

	if _, err := backend.db.ExecContext(ctx, `DELETE FROM library_memberships WHERE library_id = ? AND profile_id = ?`, string(aliceLibrary), string(bob)); err != nil {
		t.Fatalf("revoke Bob test membership: %v", err)
	}
	bobFavorites, err = backend.FavoritesForProfile(ctx, bob, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(bobFavorites) != 0 {
		t.Fatalf("revoked Bob favorites = %#v, want hidden", bobFavorites)
	}
	if _, _, err := backend.RatingForRecording(ctx, bob, aliceRecording); err == nil {
		t.Fatal("expected revoked Bob rating access to fail")
	}
	bobHistory, err := backend.RecentPlays(ctx, bob, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(bobHistory) != 0 {
		t.Fatalf("revoked Bob history = %#v, want hidden", bobHistory)
	}
}

func TestProfileStateValidation(t *testing.T) {
	ctx := context.Background()
	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer backend.Close()

	profile := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:owner")
	recording := domain.RecordingID("recording:test")
	if err := backend.CreateProfile(ctx, profile, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, profile, "Library", filepath.Join(t.TempDir(), "media")); err != nil {
		t.Fatal(err)
	}
	insertTestRecording(t, ctx, backend, recording, library, "Test")

	for _, rating := range []int{-1, 101} {
		if err := backend.SetRating(ctx, profile, recording, rating); err == nil {
			t.Fatalf("SetRating(%d) expected error", rating)
		}
	}
	if err := backend.RecordPlay(ctx, profile, recording, time.Time{}, 0); err == nil {
		t.Fatal("RecordPlay zero time expected error")
	}
	if err := backend.RecordPlay(ctx, profile, recording, time.Now(), -1); err == nil {
		t.Fatal("RecordPlay negative duration expected error")
	}
	if _, err := backend.FavoritesForProfile(ctx, profile, 0); err == nil {
		t.Fatal("FavoritesForProfile limit 0 expected error")
	}
	if _, err := backend.RecentPlays(ctx, profile, maxProfileStateResults+1); err == nil {
		t.Fatal("RecentPlays oversized limit expected error")
	}
}

func insertTestRecording(t *testing.T, ctx context.Context, backend *Backend, recordingID domain.RecordingID, libraryID domain.LibraryID, title string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := backend.db.ExecContext(ctx, `INSERT INTO recordings(recording_id, library_id, title, artist, metadata_json, added_at)
		VALUES(?, ?, ?, '', '{}', ?)`, string(recordingID), string(libraryID), title, now); err != nil {
		t.Fatalf("insert test recording: %v", err)
	}
}
