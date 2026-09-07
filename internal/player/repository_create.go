package player

import (
	"bytes"
	"context"
	"errors"
)

var ErrInvalidQueueInitialization = errors.New("invalid queue initialization")

// QueueRepositoryCreator is the optional creation boundary for a persistence
// adapter. It is separate from QueueRepository so existing adapters that only
// service already-created scopes are not forced to invent creation semantics.
type QueueRepositoryCreator interface {
	Create(ctx context.Context, scope QueueScope, queue Queue) (QueueRecord, error)
}

// InitializeStoredQueue creates the first canonical revision-zero queue record
// for a validated scope and verifies that the persistence adapter returns
// exactly the canonical value requested. It does not define user/profile
// identity, choose storage, or decide how duplicate-create conflicts are stored.
func InitializeStoredQueue(
	ctx context.Context,
	creator QueueRepositoryCreator,
	scope QueueScope,
	queue Queue,
) (QueueRecord, error) {
	if creator == nil {
		return QueueRecord{}, ErrInvalidQueueInitialization
	}
	if _, err := NewQueueScope(string(scope)); err != nil {
		return QueueRecord{}, err
	}
	if err := ctx.Err(); err != nil {
		return QueueRecord{}, err
	}

	expected, err := NewQueueRecord(queue)
	if err != nil {
		return QueueRecord{}, err
	}
	created, err := creator.Create(ctx, scope, queue)
	if err != nil {
		return QueueRecord{}, err
	}
	if created.Revision() != expected.Revision() || !bytes.Equal(created.Payload(), expected.Payload()) {
		return QueueRecord{}, ErrInvalidQueueRepositoryResult
	}
	return created, nil
}
