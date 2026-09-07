package player

import (
	"errors"
	"math"
	"testing"
)

func TestAdvanceQueueRevision(t *testing.T) {
	next, err := AdvanceQueueRevision(7, 7)
	if err != nil || next != 8 {
		t.Fatalf("next revision = %d, error = %v", next, err)
	}
}

func TestAdvanceQueueRevisionRejectsStaleWriter(t *testing.T) {
	next, err := AdvanceQueueRevision(6, 7)
	if !errors.Is(err, ErrStaleQueueRevision) || next != 7 {
		t.Fatalf("next revision = %d, error = %v", next, err)
	}
}

func TestAdvanceQueueRevisionRejectsOverflow(t *testing.T) {
	maximum := QueueRevision(math.MaxUint64)
	next, err := AdvanceQueueRevision(maximum, maximum)
	if !errors.Is(err, ErrQueueRevisionExhausted) || next != maximum {
		t.Fatalf("next revision = %d, error = %v", next, err)
	}
}
