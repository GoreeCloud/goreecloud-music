package player

import (
	"context"
	"errors"
	"testing"
)

type fakeQueueCreator struct {
	created QueueRecord
	err     error
	calls   int
}

func (f *fakeQueueCreator) Create(context.Context, QueueScope, Queue) (QueueRecord, error) {
	f.calls++
	return f.created, f.err
}

func TestInitializeStoredQueueAcceptsCanonicalRevisionZeroRecord(t *testing.T) {
	queue := NewQueue([]string{"a", "b"})
	record, err := NewQueueRecord(queue)
	if err != nil {
		t.Fatal(err)
	}
	creator := &fakeQueueCreator{created: record}

	created, err := InitializeStoredQueue(context.Background(), creator, "scope", queue)
	if err != nil || created.Revision() != 0 || creator.calls != 1 {
		t.Fatalf("revision=%d calls=%d err=%v", created.Revision(), creator.calls, err)
	}
}

func TestInitializeStoredQueueRejectsInvalidScopeBeforeCreate(t *testing.T) {
	creator := &fakeQueueCreator{}
	_, err := InitializeStoredQueue(context.Background(), creator, " scope", NewQueue([]string{"a"}))
	if !errors.Is(err, ErrInvalidQueueScope) || creator.calls != 0 {
		t.Fatalf("calls=%d err=%v", creator.calls, err)
	}
}

func TestInitializeStoredQueueHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	creator := &fakeQueueCreator{}
	_, err := InitializeStoredQueue(ctx, creator, "scope", NewQueue([]string{"a"}))
	if !errors.Is(err, context.Canceled) || creator.calls != 0 {
		t.Fatalf("calls=%d err=%v", creator.calls, err)
	}
}

func TestInitializeStoredQueuePropagatesCreatorError(t *testing.T) {
	want := errors.New("already exists")
	creator := &fakeQueueCreator{err: want}
	_, err := InitializeStoredQueue(context.Background(), creator, "scope", NewQueue([]string{"a"}))
	if !errors.Is(err, want) || creator.calls != 1 {
		t.Fatalf("calls=%d err=%v", creator.calls, err)
	}
}

func TestInitializeStoredQueueRejectsMalformedRepositoryResult(t *testing.T) {
	queue := NewQueue([]string{"a"})
	canonical, err := NewQueueRecord(queue)
	if err != nil {
		t.Fatal(err)
	}
	wrongRevision, err := canonical.Update(0, queue)
	if err != nil {
		t.Fatal(err)
	}
	creator := &fakeQueueCreator{created: wrongRevision}

	_, err = InitializeStoredQueue(context.Background(), creator, "scope", queue)
	if !errors.Is(err, ErrInvalidQueueRepositoryResult) {
		t.Fatalf("error=%v", err)
	}
}

func TestInitializeStoredQueueRejectsNilCreator(t *testing.T) {
	_, err := InitializeStoredQueue(context.Background(), nil, "scope", NewQueue([]string{"a"}))
	if !errors.Is(err, ErrInvalidQueueInitialization) {
		t.Fatalf("error=%v", err)
	}
}
