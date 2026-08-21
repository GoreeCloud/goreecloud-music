package database

import (
	"context"
	"errors"
	"testing"
)

func TestOpenPostgresRequiresURL(t *testing.T) {
	_, err := OpenPostgres(context.Background(), "   ")
	if !errors.Is(err, ErrMissingDatabaseURL) {
		t.Fatalf("expected ErrMissingDatabaseURL, got %v", err)
	}
}
