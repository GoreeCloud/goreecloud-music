package player

import (
	"errors"
	"strings"
	"testing"
)

func TestQueueScopeAcceptsOpaqueExactValue(t *testing.T) {
	scope, err := NewQueueScope("profile:primary/device:phone")
	if err != nil {
		t.Fatal(err)
	}
	if scope != QueueScope("profile:primary/device:phone") {
		t.Fatalf("scope = %q", scope)
	}
}

func TestQueueScopeRejectsBlankWhitespaceAndNormalization(t *testing.T) {
	for _, value := range []string{"", " ", " queue", "queue ", "\tqueue"} {
		if _, err := NewQueueScope(value); !errors.Is(err, ErrInvalidQueueScope) {
			t.Fatalf("scope %q error = %v, want invalid scope", value, err)
		}
	}
}

func TestQueueScopeRejectsOversizedNamespace(t *testing.T) {
	value := strings.Repeat("a", MaxQueueScopeLength+1)
	if _, err := NewQueueScope(value); !errors.Is(err, ErrInvalidQueueScope) {
		t.Fatalf("error = %v, want invalid scope", err)
	}
}

func TestQueueScopeAcceptsMaximumLength(t *testing.T) {
	value := strings.Repeat("a", MaxQueueScopeLength)
	scope, err := NewQueueScope(value)
	if err != nil {
		t.Fatal(err)
	}
	if string(scope) != value {
		t.Fatal("maximum-length queue scope changed")
	}
}
