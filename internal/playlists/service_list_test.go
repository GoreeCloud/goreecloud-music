package playlists

import (
	"context"
	"errors"
	"testing"
)

type listingRepository struct {
	fakeRepository
	records []Record
	listErr error
	lists   int
}

func (r *listingRepository) List(ctx context.Context, userID string) ([]Record, error) {
	r.lists++
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.listErr != nil {
		return nil, r.listErr
	}
	return append([]Record(nil), r.records...), nil
}

func TestPlaylistServiceListReturnsDeterministicMetadata(t *testing.T) {
	first, _ := New("user-1", "z-playlist", "Zulu")
	_ = first.Append("track-1")
	_ = first.Append("track-2")
	second, _ := New("user-1", "a-playlist", "Alpha")
	_ = second.Append("track-3")
	firstRecord, _ := NewRecord(first)
	secondRecord, _ := NewRecord(second)
	repository := &listingRepository{records: []Record{firstRecord, secondRecord}}
	service, _ := NewService(repository)

	entries, err := service.List(context.Background(), "user-1")
	if err != nil || repository.lists != 1 {
		t.Fatalf("entries=%v lists=%d err=%v", entries, repository.lists, err)
	}
	if len(entries) != 2 || entries[0].ID != "a-playlist" || entries[0].Name != "Alpha" || entries[0].TrackCount != 1 || entries[1].ID != "z-playlist" || entries[1].TrackCount != 2 {
		t.Fatalf("entries=%+v", entries)
	}
}

func TestPlaylistServiceListRejectsWrongUserDuplicateAndMalformedRecords(t *testing.T) {
	wrong, _ := New("user-2", "playlist-1", "Wrong")
	wrongRecord, _ := NewRecord(wrong)
	repository := &listingRepository{records: []Record{wrongRecord}}
	service, _ := NewService(repository)
	if _, err := service.List(context.Background(), "user-1"); !errors.Is(err, ErrInvalidRepositoryResult) {
		t.Fatalf("wrong-user error=%v", err)
	}

	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	repository.records = []Record{record, record}
	if _, err := service.List(context.Background(), "user-1"); !errors.Is(err, ErrInvalidRepositoryResult) {
		t.Fatalf("duplicate error=%v", err)
	}

	malformed := Record{revision: 0, payload: []byte("not-json")}
	repository.records = []Record{malformed}
	if _, err := service.List(context.Background(), "user-1"); !errors.Is(err, ErrInvalidRepositoryResult) {
		t.Fatalf("malformed error=%v", err)
	}
}

func TestPlaylistServiceListPropagatesUnsupportedRepositoryErrorAndCancellation(t *testing.T) {
	service, _ := NewService(&fakeRepository{})
	if _, err := service.List(context.Background(), "user-1"); !errors.Is(err, ErrListingUnsupported) {
		t.Fatalf("unsupported error=%v", err)
	}

	want := errors.New("list failed")
	repository := &listingRepository{listErr: want}
	service, _ = NewService(repository)
	if _, err := service.List(context.Background(), "user-1"); !errors.Is(err, want) {
		t.Fatalf("list error=%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.List(ctx, "user-1"); !errors.Is(err, context.Canceled) || repository.lists != 1 {
		t.Fatalf("lists=%d cancellation=%v", repository.lists, err)
	}

	if _, err := service.List(context.Background(), " user-1 "); !errors.Is(err, ErrInvalidService) {
		t.Fatalf("invalid user error=%v", err)
	}
}
