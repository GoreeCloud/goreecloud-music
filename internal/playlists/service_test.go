package playlists

import (
	"context"
	"errors"
	"testing"
)

func TestPlaylistServiceCreateLoadAndMutate(t *testing.T) {
	repository := &fakeRepository{}
	service, err := NewService(repository)
	if err != nil {
		t.Fatal(err)
	}

	record, err := service.Create(context.Background(), "user-1", "playlist-1", "Road Trip")
	if err != nil || record.Revision() != 0 || repository.creates != 1 {
		t.Fatalf("record=%+v creates=%d err=%v", record, repository.creates, err)
	}

	if _, err := service.Append(context.Background(), "user-1", "playlist-1", "track-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Append(context.Background(), "user-1", "playlist-1", "track-b"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Insert(context.Background(), "user-1", "playlist-1", 1, "track-x"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Move(context.Background(), "user-1", "playlist-1", 2, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Rename(context.Background(), "user-1", "playlist-1", "Updated"); err != nil {
		t.Fatal(err)
	}

	_, playlist, err := service.Load(context.Background(), "user-1", "playlist-1")
	if err != nil || playlist.Name() != "Updated" {
		t.Fatalf("playlist=%+v err=%v", playlist, err)
	}
	tracks := playlist.TrackIDs()
	if len(tracks) != 3 || tracks[0] != "track-b" || tracks[1] != "track-a" || tracks[2] != "track-x" {
		t.Fatalf("tracks=%v", tracks)
	}

	record, removed, err := service.RemoveAt(context.Background(), "user-1", "playlist-1", 1)
	if err != nil || removed != "track-a" || record.Revision() != 6 {
		t.Fatalf("removed=%q revision=%d err=%v", removed, record.Revision(), err)
	}
}

func TestPlaylistServicePropagatesDomainAndRepositoryErrors(t *testing.T) {
	repository := &fakeRepository{}
	service, err := NewService(repository)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), " user ", "playlist-1", "Mix"); !errors.Is(err, ErrInvalidPlaylistState) {
		t.Fatalf("create error=%v", err)
	}

	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	repository.record = record
	repository.found = true
	if _, err := service.Append(context.Background(), "user-1", "playlist-1", " bad "); !errors.Is(err, ErrInvalidPlaylistState) {
		t.Fatalf("append error=%v", err)
	}
	if repository.saves != 0 {
		t.Fatalf("unexpected saves=%d", repository.saves)
	}

	want := errors.New("save failed")
	repository.saveErr = want
	if _, err := service.Rename(context.Background(), "user-1", "playlist-1", "Changed"); !errors.Is(err, want) {
		t.Fatalf("save error=%v", err)
	}
}

func TestPlaylistServiceLoadValidatesIdentityAndNotFound(t *testing.T) {
	service, err := NewService(&fakeRepository{})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.Load(context.Background(), "user-1", "playlist-1"); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("not found error=%v", err)
	}

	wrong, _ := New("user-2", "playlist-2", "Wrong")
	wrongRecord, _ := NewRecord(wrong)
	repository := &fakeRepository{record: wrongRecord, found: true}
	service, _ = NewService(repository)
	if _, _, err := service.Load(context.Background(), "user-1", "playlist-1"); !errors.Is(err, ErrInvalidRepositoryResult) {
		t.Fatalf("identity error=%v", err)
	}
}

func TestPlaylistServiceHonorsCanceledContextAndInvalidZeroService(t *testing.T) {
	repository := &fakeRepository{}
	service, _ := NewService(repository)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Create(ctx, "user-1", "playlist-1", "Mix"); !errors.Is(err, context.Canceled) {
		t.Fatalf("create context error=%v", err)
	}

	var zero Service
	if _, _, err := zero.Load(context.Background(), "user-1", "playlist-1"); !errors.Is(err, ErrInvalidService) {
		t.Fatalf("zero load error=%v", err)
	}
	if _, err := zero.Append(context.Background(), "user-1", "playlist-1", "track-1"); !errors.Is(err, ErrInvalidService) {
		t.Fatalf("zero append error=%v", err)
	}
}
