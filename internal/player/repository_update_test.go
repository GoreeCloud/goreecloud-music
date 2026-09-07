package player

import (
	"context"
	"errors"
	"testing"
)

type fakeQueueRepository struct {
	record  QueueRecord
	found   bool
	loadErr error
	saveErr error
	saves   int
}

func (f *fakeQueueRepository) Load(context.Context, QueueScope) (QueueRecord, bool, error) {
	return f.record, f.found, f.loadErr
}

func (f *fakeQueueRepository) Save(
	_ context.Context,
	_ QueueScope,
	expected QueueRevision,
	queue Queue,
) (QueueRecord, error) {
	f.saves++
	if f.saveErr != nil {
		return f.record, f.saveErr
	}
	next, err := f.record.Update(expected, queue)
	if err != nil {
		return f.record, err
	}
	f.record = next
	return next, nil
}

func TestMutateStoredQueueSavesAgainstLoadedRevision(t *testing.T) {
	record, err := NewQueueRecord(NewQueue([]string{"a"}))
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeQueueRepository{record: record, found: true}

	next, err := MutateStoredQueue(context.Background(), repository, "scope", func(queue *Queue) error {
		return queue.Append("b")
	})
	if err != nil {
		t.Fatal(err)
	}
	restored, err := next.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if next.Revision() != 1 || len(restored.TrackIDs) != 2 || restored.TrackIDs[1] != "b" || repository.saves != 1 {
		t.Fatalf("next revision=%d queue=%+v saves=%d", next.Revision(), restored, repository.saves)
	}
}

func TestMutateStoredQueueReturnsNotFound(t *testing.T) {
	repository := &fakeQueueRepository{}
	_, err := MutateStoredQueue(context.Background(), repository, "scope", func(*Queue) error { return nil })
	if !errors.Is(err, ErrQueueRecordNotFound) {
		t.Fatalf("error = %v, want queue record not found", err)
	}
}

func TestMutateStoredQueueDoesNotSaveMutationFailure(t *testing.T) {
	record, err := NewQueueRecord(NewQueue([]string{"a"}))
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeQueueRepository{record: record, found: true}
	want := errors.New("mutation failed")

	returned, err := MutateStoredQueue(context.Background(), repository, "scope", func(*Queue) error { return want })
	if !errors.Is(err, want) || returned.Revision() != record.Revision() || repository.saves != 0 {
		t.Fatalf("returned revision=%d error=%v saves=%d", returned.Revision(), err, repository.saves)
	}
}

func TestMutateStoredQueuePropagatesStaleWriter(t *testing.T) {
	record, err := NewQueueRecord(NewQueue([]string{"a"}))
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeQueueRepository{record: record, found: true, saveErr: ErrStaleQueueRevision}

	_, err = MutateStoredQueue(context.Background(), repository, "scope", func(queue *Queue) error {
		return queue.Append("b")
	})
	if !errors.Is(err, ErrStaleQueueRevision) {
		t.Fatalf("error = %v, want stale queue revision", err)
	}
}

func TestMutateStoredQueueHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repository := &fakeQueueRepository{}

	_, err := MutateStoredQueue(ctx, repository, "scope", func(*Queue) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
}

func TestMutateStoredQueueRejectsNilRepository(t *testing.T) {
	_, err := MutateStoredQueue(context.Background(), nil, "scope", func(*Queue) error { return nil })
	if !errors.Is(err, ErrInvalidQueueMutation) {
		t.Fatalf("error = %v, want invalid queue mutation", err)
	}
}
