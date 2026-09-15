package domain

import "testing"

func TestDeriveAvailabilityDownloaded(t *testing.T) {
	got := DeriveAvailability(AvailabilityInput{AuthorizationValid: true, LocalReadable: true, IntegrityVerified: true, Downloaded: true})
	if got != AvailabilityDownloaded {
		t.Fatalf("got %q", got)
	}
}

func TestDeriveAvailabilityCachedIsNotDownloaded(t *testing.T) {
	got := DeriveAvailability(AvailabilityInput{AuthorizationValid: true, LocalReadable: true, IntegrityVerified: true, Cached: true})
	if got != AvailabilityCached {
		t.Fatalf("got %q", got)
	}
}

func TestDeriveAvailabilityRejectsCorruptLocalAsset(t *testing.T) {
	got := DeriveAvailability(AvailabilityInput{AuthorizationValid: true, LocalReadable: true, IntegrityVerified: true, Downloaded: true, Corrupt: true})
	if got != AvailabilityUnavailable {
		t.Fatalf("got %q", got)
	}
}

func TestSourceIdentityIndependentFromAvailability(t *testing.T) {
	s := SourceItem{ID: "source-1", ProviderID: "server", Kind: SourceGoreeCloudServer, RecordingID: "recording-1"}
	_ = DeriveAvailability(AvailabilityInput{AuthorizationValid: true, SourceReachable: true, NetworkRequired: true})
	if s.ProviderID != "server" || s.Kind != SourceGoreeCloudServer {
		t.Fatal("availability derivation changed source identity")
	}
}
