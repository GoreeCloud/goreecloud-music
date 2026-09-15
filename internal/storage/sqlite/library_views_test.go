package sqlitestore

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func TestRecentlyAddedForProfileAuthorizationOrderingAndRevocation(t *testing.T) {
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

	base := time.Date(2026, 9, 15, 18, 0, 0, 0, time.UTC)
	oldRecording := domain.RecordingID("recording:alice-old")
	newRecording := domain.RecordingID("recording:alice-new")
	bobRecording := domain.RecordingID("recording:bob-secret")
	insertRecentlyAddedRecording(t, ctx, backend, oldRecording, aliceLibrary, "Older", "Artist A", base)
	insertRecentlyAddedRecording(t, ctx, backend, newRecording, aliceLibrary, "Newer", "Artist B", base.Add(time.Hour))
	insertRecentlyAddedRecording(t, ctx, backend, bobRecording, bobLibrary, "Bob Secret", "Artist C", base.Add(2*time.Hour))

	aliceItems, err := backend.RecentlyAddedForProfile(ctx, alice, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(aliceItems) != 2 {
		t.Fatalf("Alice recently added len = %d, want 2", len(aliceItems))
	}
	if aliceItems[0].RecordingID != newRecording || aliceItems[1].RecordingID != oldRecording {
		t.Fatalf("Alice ordering = %#v", aliceItems)
	}
	for _, item := range aliceItems {
		if item.LibraryID != aliceLibrary {
			t.Fatalf("Alice saw inaccessible library item: %#v", item)
		}
	}

	bobItems, err := backend.RecentlyAddedForProfile(ctx, bob, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(bobItems) != 1 || bobItems[0].RecordingID != bobRecording {
		t.Fatalf("Bob recently added = %#v", bobItems)
	}

	if err := backend.SetLibraryPermission(ctx, aliceLibrary, bob, LibraryPermissionRead); err != nil {
		t.Fatalf("grant Bob read access: %v", err)
	}
	bobItems, err = backend.RecentlyAddedForProfile(ctx, bob, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(bobItems) != 3 {
		t.Fatalf("Bob recently added after grant len = %d, want 3", len(bobItems))
	}
	if bobItems[0].RecordingID != bobRecording || bobItems[1].RecordingID != newRecording || bobItems[2].RecordingID != oldRecording {
		t.Fatalf("Bob ordering after grant = %#v", bobItems)
	}

	if _, err := backend.db.ExecContext(ctx, `DELETE FROM library_memberships WHERE library_id = ? AND profile_id = ?`, string(aliceLibrary), string(bob)); err != nil {
		t.Fatalf("revoke Bob test membership: %v", err)
	}
	bobItems, err = backend.RecentlyAddedForProfile(ctx, bob, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(bobItems) != 1 || bobItems[0].RecordingID != bobRecording {
		t.Fatalf("Bob recently added after revoke = %#v", bobItems)
	}
}

func TestRecentlyAddedForProfileValidationAndLimit(t *testing.T) {
	ctx := context.Background()
	backend, err := Open(ctx, filepath.Join(t.TempDir(), "music.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer backend.Close()

	profile := domain.ProfileID("profile:owner")
	library := domain.LibraryID("library:owner")
	if err := backend.CreateProfile(ctx, profile, "Owner"); err != nil {
		t.Fatal(err)
	}
	if err := backend.CreateLibrary(ctx, library, profile, "Library", filepath.Join(t.TempDir(), "media")); err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 9, 15, 18, 0, 0, 0, time.UTC)
	insertRecentlyAddedRecording(t, ctx, backend, domain.RecordingID("recording:first"), library, "First", "Artist", base)
	insertRecentlyAddedRecording(t, ctx, backend, domain.RecordingID("recording:second"), library, "Second", "Artist", base.Add(time.Minute))

	items, err := backend.RecentlyAddedForProfile(ctx, profile, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].RecordingID != domain.RecordingID("recording:second") {
		t.Fatalf("limited recently added = %#v", items)
	}
	if _, err := backend.RecentlyAddedForProfile(ctx, profile, 0); err == nil {
		t.Fatal("RecentlyAddedForProfile limit 0 expected error")
	}
	if _, err := backend.RecentlyAddedForProfile(ctx, profile, maxProfileStateResults+1); err == nil {
		t.Fatal("RecentlyAddedForProfile oversized limit expected error")
	}
	if _, err := backend.RecentlyAddedForProfile(ctx, domain.ProfileID("profile/invalid"), 1); err == nil {
		t.Fatal("RecentlyAddedForProfile invalid profile expected error")
	}
}

func insertRecentlyAddedRecording(t *testing.T, ctx context.Context, backend *Backend, recordingID domain.RecordingID, libraryID domain.LibraryID, title, artist string, addedAt time.Time) {
	t.Helper()
	if _, err := backend.db.ExecContext(ctx, `INSERT INTO recordings(recording_id, library_id, title, artist, metadata_json, added_at)
		VALUES(?, ?, ?, ?, '{}', ?)`, string(recordingID), string(libraryID), title, artist, addedAt.UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("insert recently added recording: %v", err)
	}
}
