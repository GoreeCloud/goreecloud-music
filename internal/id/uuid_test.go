package id

import "testing"

func TestNewUUID(t *testing.T) {
	value, err := NewUUID()
	if err != nil {
		t.Fatalf("NewUUID returned error: %v", err)
	}
	if len(value) != 36 {
		t.Fatalf("expected 36-character UUID, got %q", value)
	}
	if value[14] != '4' {
		t.Fatalf("expected version 4 UUID, got %q", value)
	}
	if value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		t.Fatalf("expected RFC 4122 separators, got %q", value)
	}
}
