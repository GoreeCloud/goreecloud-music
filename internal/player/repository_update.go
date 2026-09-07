package player

import (
	"bytes"
	"context"
	"errors"
)

var (
	ErrQueueRecordNotFound          = errors.New("queue record not found")
	ErrInvalidQueueMutation         = errors.New("invalid queue mutation")
	ErrInvalidQueueRepositoryResult = errors.New("invalid queue repository result")
)

// MutateStoredQueue performs one optimistic load-mutate-save cycle. Repository
// implementations remain responsible for atomically enforcing the expected
// revision passed to Save. The player core additionally validates the mutated
// queue before Save and verifies that a successful repository response contains
// exactly the next revision and canonical payload that were requested.
func MutateStoredQueue(
	ctx context.Context,
	repository QueueRepository,
	scope QueueScope,
	mutate func(*Queue) error,
) (QueueRecord, error) {
	if repository == nil || mutate == nil {
		return QueueRecord{}, ErrInvalidQueueMutation
	}
	if _, err := NewQueueScope(string(scope)); err != nil {
		return QueueRecord{}, err
	}
	if err := ctx.Err(); err != nil {
		return QueueRecord{}, err
	}

	record, found, err := repository.Load(ctx, scope)
	if err != nil {
		return QueueRecord{}, err
	}
	if !found {
		return QueueRecord{}, ErrQueueRecordNotFound
	}
	queue, err := record.Restore()
	if err != nil {
		return record, err
	}
	if err := mutate(&queue); err != nil {
		return record, err
	}
	if err := ctx.Err(); err != nil {
		return record, err
	}

	expectedPayload, err := EncodeQueueSnapshot(queue)
	if err != nil {
		return record, err
	}
	expectedRevision, err := AdvanceQueueRevision(record.Revision(), record.Revision())
	if err != nil {
		return record, err
	}

	saved, err := repository.Save(ctx, scope, record.Revision(), queue)
	if err != nil {
		return record, err
	}
	if saved.Revision() != expectedRevision || !bytes.Equal(saved.Payload(), expectedPayload) {
		return record, ErrInvalidQueueRepositoryResult
	}
	return saved, nil
}
