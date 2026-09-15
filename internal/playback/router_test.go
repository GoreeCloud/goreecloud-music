package playback

import (
	"errors"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func candidate(provider string, kind domain.SourceKind, asset string, match domain.MatchConfidence, availability domain.AvailabilityState) Candidate {
	return Candidate{ProviderID: provider, Kind: kind, RecordingID: "recording-1", SourceItemID: domain.SourceItemID(provider + "-item"), AssetID: domain.PlayableAssetID(asset), Match: match, Availability: availability, Authorized: true, RequiresNetwork: availability == domain.AvailabilityOnline}
}

func TestDownloadedExactCopyWins(t *testing.T) {
	got, err := Select(Request{RecordingID: "recording-1", NetworkAvailable: true}, []Candidate{
		candidate("server", domain.SourceGoreeCloudServer, "server-stream", domain.MatchExact, domain.AvailabilityOnline),
		candidate("server", domain.SourceGoreeCloudServer, "local-copy", domain.MatchExact, domain.AvailabilityDownloaded),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Candidate.AssetID != "local-copy" || got.Reason != ReasonDownloadedCopy {
		t.Fatalf("unexpected decision: %+v", got)
	}
}

func TestNonExactIsNotSilentlySubstituted(t *testing.T) {
	_, err := Select(Request{RecordingID: "recording-1", NetworkAvailable: true}, []Candidate{
		candidate("external", domain.SourceExternal, "remaster", domain.MatchAlternate, domain.AvailabilityOnline),
	})
	if !errors.Is(err, ErrNoPlayableRoute) {
		t.Fatalf("expected no route, got %v", err)
	}
}

func TestExplicitNonExactPermissionAllowsRoute(t *testing.T) {
	got, err := Select(Request{RecordingID: "recording-1", NetworkAvailable: true, AllowNonExact: true}, []Candidate{
		candidate("external", domain.SourceExternal, "alternate", domain.MatchAlternate, domain.AvailabilityOnline),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Candidate.AssetID != "alternate" {
		t.Fatalf("unexpected asset: %s", got.Candidate.AssetID)
	}
}

func TestUnauthorizedCandidateSkipped(t *testing.T) {
	bad := candidate("external", domain.SourceExternal, "external", domain.MatchExact, domain.AvailabilityDownloaded)
	bad.Authorized = false
	good := candidate("server", domain.SourceGoreeCloudServer, "server", domain.MatchExact, domain.AvailabilityOnline)
	got, err := Select(Request{RecordingID: "recording-1", NetworkAvailable: true}, []Candidate{bad, good})
	if err != nil {
		t.Fatal(err)
	}
	if got.Candidate.AssetID != "server" {
		t.Fatalf("unexpected asset: %s", got.Candidate.AssetID)
	}
}

func TestProviderOutageDoesNotBlockOfflineSource(t *testing.T) {
	online := candidate("external", domain.SourceExternal, "external", domain.MatchExact, domain.AvailabilityOnline)
	offline := candidate("server", domain.SourceGoreeCloudServer, "download", domain.MatchExact, domain.AvailabilityDownloaded)
	got, err := Select(Request{RecordingID: "recording-1", NetworkAvailable: false}, []Candidate{online, offline})
	if err != nil {
		t.Fatal(err)
	}
	if got.Candidate.AssetID != "download" {
		t.Fatalf("unexpected asset: %s", got.Candidate.AssetID)
	}
}

func TestAskWhenMultipleProviders(t *testing.T) {
	_, err := Select(Request{RecordingID: "recording-1", NetworkAvailable: true, AskWhenMultiple: true}, []Candidate{
		candidate("server", domain.SourceGoreeCloudServer, "server", domain.MatchExact, domain.AvailabilityOnline),
		candidate("external", domain.SourceExternal, "external", domain.MatchExact, domain.AvailabilityOnline),
	})
	if !errors.Is(err, ErrUserChoiceRequired) {
		t.Fatalf("expected user choice, got %v", err)
	}
}
