package streaming

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrUnsatisfiableRange = errors.New("unsatisfiable byte range")

type ByteRange struct {
	Start int64
	End   int64
}

func (r ByteRange) Length() int64 { return r.End - r.Start + 1 }

// ParseSingleRange parses one RFC 7233-style bytes range for a resource of size bytes.
// Multiple ranges are intentionally rejected for the initial Resonance streaming contract.
func ParseSingleRange(header string, size int64) (ByteRange, error) {
	if size <= 0 {
		return ByteRange{}, ErrUnsatisfiableRange
	}
	if header == "" {
		return ByteRange{Start: 0, End: size - 1}, nil
	}
	if !strings.HasPrefix(header, "bytes=") {
		return ByteRange{}, fmt.Errorf("invalid range unit")
	}
	spec := strings.TrimSpace(strings.TrimPrefix(header, "bytes="))
	if spec == "" || strings.Contains(spec, ",") {
		return ByteRange{}, fmt.Errorf("exactly one byte range is required")
	}
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return ByteRange{}, fmt.Errorf("invalid byte range")
	}

	if parts[0] == "" {
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffix <= 0 {
			return ByteRange{}, fmt.Errorf("invalid suffix range")
		}
		if suffix > size {
			suffix = size
		}
		return ByteRange{Start: size - suffix, End: size - 1}, nil
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= size {
		return ByteRange{}, ErrUnsatisfiableRange
	}
	end := size - 1
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start {
			return ByteRange{}, ErrUnsatisfiableRange
		}
		if end >= size {
			end = size - 1
		}
	}
	return ByteRange{Start: start, End: end}, nil
}
