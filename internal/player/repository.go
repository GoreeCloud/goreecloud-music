package player

import (
	"context"
	"errors"
	"strings"
)

const MaxQueueScopeLength = 256

var ErrInvalidQueueScope = errors.New("invalid queue scope")

// QueueScope is an opaque persistence namespace. The player core deliberately
// does not define whether a scope maps to a user, profile, device, session, or
// another identity-owned namespace.
type QueueScope string

func NewQueueScope(value string) (QueueScope, error) {
	if value == "" || len(value) > MaxQueueScopeLength || strings.TrimSpace(value) != value {
		return "", ErrInvalidQueueScope
	}
	return QueueScope(value), nil
}

// QueueRepository is the persistence boundary for Resonance queue state.
// Implementations choose the actual database or storage service and must honor
// QueueRecord revision semantics when saving.
type QueueRepository interface {
	Load(ctx context.Context, scope QueueScope) (QueueRecord, bool, error)
	Save(ctx context.Context, scope QueueScope, expected QueueRevision, queue Queue) (QueueRecord, error)
}
