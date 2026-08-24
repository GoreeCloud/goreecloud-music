package streaming

import (
	"errors"
	"testing"
)

func TestParseSingleRange(t *testing.T) {
	tests := []struct {
		name   string
		header string
		size   int64
		want   ByteRange
	}{
		{"full", "", 100, ByteRange{0, 99}},
		{"bounded", "bytes=10-19", 100, ByteRange{10, 19}},
		{"open ended", "bytes=90-", 100, ByteRange{90, 99}},
		{"suffix", "bytes=-10", 100, ByteRange{90, 99}},
		{"clamped", "bytes=90-200", 100, ByteRange{90, 99}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSingleRange(tt.header, tt.size)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
}

func TestParseSingleRangeRejectsUnsatisfiable(t *testing.T) {
	_, err := ParseSingleRange("bytes=100-", 100)
	if !errors.Is(err, ErrUnsatisfiableRange) {
		t.Fatalf("expected ErrUnsatisfiableRange, got %v", err)
	}
}

func TestParseSingleRangeRejectsMultipleRanges(t *testing.T) {
	if _, err := ParseSingleRange("bytes=0-1,4-5", 100); err == nil {
		t.Fatal("expected multiple ranges to be rejected")
	}
}
