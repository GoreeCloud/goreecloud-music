package playlists

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	record       Record
	found        bool
	loadErr      error
	createErr    error
	saveErr      error
	deleteErr    error
	createResult *Record
	saveResult   *Record
	creates      int
	saves        int
	deletes      int
	deleteResult bool
}

func (f *fakeRepository) Load(context.Context, string, string) (Record, bool, error) {
	return f.record, f.found, f.loadErr
}

func (f *fakeRepository) Create(_ context.Context, playlist Playlist) (Record, error) {
	f.creates++
	if f.createErr != nil {
		return Record{}, f.createErr
	}
	if f.createResult != nil {
		return *f.createResult, nil
	}
	record, err := NewRecord(playlist)
	if err != nil {
		return Record{}, err
	}
	f.record = record
	f.found = true
	return record, nil
}

func (f *fakeRepository) Save(_ context.Context, expected Revision, playlist Playlist) (Record, error) {
	f.saves++
	if f.saveErr != nil {
		return f.record, f.saveErr
	}
	if f.saveResult != nil {
		return *f.saveResult, nil
	}
	next, err := f.record.Update(expected, playlist)
	if err != nil {
		return f.record, err
	}
	f.record = next
	return next, nil
}

func (f *fakeRepository) Delete(_ context.Context, _, _ string, _ Revision) (bool, error) {
	f.deletes++
	if f.deleteErr != nil {
		return false, f.deleteErr
	}
	if f.deleteResult {
		f.found = false
		return true, nil
	}
	return false, nil
}

func TestInitializeStoredValidatesRepositoryResult(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	repository := &fakeRepository{}
	record, err := InitializeStored(context.Background(), repository, playlist)
	if err != nil || record.Revision() != 0 || repository.creates != 1 {
		t.Fatalf("record=%+v creates=%d err=%v", record, repository.creates, err)
	}

	wrong, _ := New("user-2", "playlist-2", "Mix")
	wrongRecord, _ := NewRecord(wrong)
	repository = &fakeRepository{createResult: &wrongRecord}
	if _, err := InitializeStored(context.Background(), repository, playlist); !errors.Is(err, ErrInvalidRepositoryResult) {
		t.Fatalf("error=%v", err)
	}
}

func TestMutateStoredSavesCanonicalNextRevision(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	repository := &fakeRepository{record: record, found: true}

	next, err := MutateStored(context.Background(), repository, "user-1", "playlist-1", func(current Playlist) (Playlist, error) {
		if err := current.Append("a"); err != nil {
			return Playlist{}, err
		}
		return current, nil
	})
	if err != nil || next.Revision() != 1 || repository.saves != 1 {
		t.Fatalf("next=%+v saves=%d err=%v", next, repository.saves, err)
	}
	restored, err := next.Restore()
	if err != nil || restored.Len() != 1 || restored.TrackIDs()[0] != "a" {
		t.Fatalf("tracks=%v err=%v", restored.TrackIDs(), err)
	}
}

func TestMutateStoredRejectsWrongLoadAndSaveResults(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	wrong, _ := New("user-2", "playlist-2", "Mix")
	wrongRecord, _ := NewRecord(wrong)

	repository := &fakeRepository{record: wrongRecord, found: true}
	if _, err := MutateStored(context.Background(), repository, "user-1", "playlist-1", func(current Playlist) (Playlist, error) {
		return current, nil
	}); !errors.Is(err, ErrInvalidRepositoryResult) {
		t.Fatalf("wrong load error=%v", err)
	}

	repository = &fakeRepository{record: record, found: true, saveResult: &wrongRecord}
	if _, err := MutateStored(context.Background(), repository, "user-1", "playlist-1", func(current Playlist) (Playlist, error) {
		_ = current.Append("a")
		return current, nil
	}); !errors.Is(err, ErrInvalidRepositoryResult) {
		t.Fatalf("wrong save error=%v", err)
	}
}

func TestMutateStoredPropagatesNotFoundMutationAndContext(t *testing.T) {
	repository := &fakeRepository{}
	if _, err := MutateStored(context.Background(), repository, "user-1", "playlist-1", func(current Playlist) (Playlist, error) {
		return current, nil
	}); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("not found error=%v", err)
	}

	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	repository = &fakeRepository{record: record, found: true}
	want := errors.New("mutation failed")
	if _, err := MutateStored(context.Background(), repository, "user-1", "playlist-1", func(Playlist) (Playlist, error) {
		return Playlist{}, want
	}); !errors.Is(err, want) || repository.saves != 0 {
		t.Fatalf("error=%v saves=%d", err, repository.saves)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := MutateStored(ctx, repository, "user-1", "playlist-1", func(current Playlist) (Playlist, error) {
		return current, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("context error=%v", err)
	}
}

func TestDeleteStoredUsesLoadedRevisionAndReturnsDeletedRecord(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	repository := &fakeRepository{record: record, found: true, deleteResult: true}
	deleted, err := DeleteStored(context.Background(), repository, "user-1", "playlist-1")
	if err != nil || repository.deletes != 1 || !sameRecord(deleted, record) || repository.found {
		t.Fatalf("deleted=%+v deletes=%d found=%v err=%v", deleted, repository.deletes, repository.found, err)
	}
}

func TestDeleteStoredFailsClosedOnNotFoundWrongIdentityAndRepositoryResult(t *testing.T) {
	if _, err := DeleteStored(context.Background(), &fakeRepository{}, "user-1", "playlist-1"); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("not found error=%v", err)
	}

	wrong, _ := New("user-2", "playlist-2", "Mix")
	wrongRecord, _ := NewRecord(wrong)
	repository := &fakeRepository{record: wrongRecord, found: true, deleteResult: true}
	if _, err := DeleteStored(context.Background(), repository, "user-1", "playlist-1"); !errors.Is(err, ErrInvalidRepositoryResult) || repository.deletes != 0 {
		t.Fatalf("wrong identity error=%v deletes=%d", err, repository.deletes)
	}

	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	repository = &fakeRepository{record: record, found: true}
	if _, err := DeleteStored(context.Background(), repository, "user-1", "playlist-1"); !errors.Is(err, ErrInvalidRepositoryResult) || repository.deletes != 1 {
		t.Fatalf("false result error=%v deletes=%d", err, repository.deletes)
	}
}

func TestDeleteStoredPropagatesRepositoryAndContextErrors(t *testing.T) {
	playlist, _ := New("user-1", "playlist-1", "Mix")
	record, _ := NewRecord(playlist)
	want := errors.New("delete failed")
	repository := &fakeRepository{record: record, found: true, deleteErr: want}
	if _, err := DeleteStored(context.Background(), repository, "user-1", "playlist-1"); !errors.Is(err, want) {
		t.Fatalf("delete error=%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := DeleteStored(ctx, repository, "user-1", "playlist-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("context error=%v", err)
	}
}
