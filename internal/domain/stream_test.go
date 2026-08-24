package domain

import "testing"

func TestStreamRequestValidate(t *testing.T) {
	end := int64(1023)
	tests := []struct {
		name    string
		request StreamRequest
		wantErr bool
	}{
		{name: "valid full stream", request: StreamRequest{ActorID: "user", LibraryID: "library", TrackID: "track"}},
		{name: "valid range", request: StreamRequest{ActorID: "user", LibraryID: "library", TrackID: "track", Range: &ByteRange{Start: 0, End: &end}}},
		{name: "missing actor", request: StreamRequest{LibraryID: "library", TrackID: "track"}, wantErr: true},
		{name: "negative range", request: StreamRequest{ActorID: "user", LibraryID: "library", TrackID: "track", Range: &ByteRange{Start: -1}}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if test.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}
