package player

import (
	"errors"
	"math"
)

type QueueRevision uint64

var (
	ErrStaleQueueRevision     = errors.New("stale queue revision")
	ErrQueueRevisionExhausted = errors.New("queue revision exhausted")
)

// AdvanceQueueRevision is a storage-neutral optimistic-concurrency boundary.
// A persistence adapter may commit only when the caller's expected revision
// matches the currently stored revision.
func AdvanceQueueRevision(expected, current QueueRevision) (QueueRevision, error) {
	if expected != current {
		return current, ErrStaleQueueRevision
	}
	if current == QueueRevision(math.MaxUint64) {
		return current, ErrQueueRevisionExhausted
	}
	return current + 1, nil
}
