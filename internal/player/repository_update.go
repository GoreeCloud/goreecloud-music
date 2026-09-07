package player

import (
	"context"
	"errors"
)

var (
	ErrQueueRecordNotFound  = errors.New("queue record not found")
	ErrInvalidQueueMutation = errors.New("invalid queue mutation")
)

// MutateStoredQueue performs one optimistic load-mutate-save cycle. Repository
// implementations remain responsible for atomically enforcing the expected
// revision passed to Save. Stale-write errors are returned unchanged.
func MutateStoredQueue(
	ctx context.Context,
	repository QueueRepository,
	scope QueueScope,
	mutate func(*Queue) error,
) (QueueRecord, error) {
	if repository == nil || mutate == nil {
		return QueueRecord{}, ErrInvalidQueueMutation
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
	return repository.Save(ctx, scope, record.Revision(), queue)
}
