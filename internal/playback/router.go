package playback

import (
	"errors"
	"sort"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

var (
	ErrNoPlayableRoute    = errors.New("no authorized playable route")
	ErrUserChoiceRequired = errors.New("multiple playable sources require user choice")
)

type SourcePreference string

const (
	PreferDefault          SourcePreference = "default"
	PreferOffline          SourcePreference = "prefer-offline"
	PreferGoreeCloudServer SourcePreference = "prefer-goreecloud-server"
)

type RouteReason string

const (
	ReasonRequestedProvider RouteReason = "explicit-provider-source"
	ReasonDownloadedCopy    RouteReason = "downloaded-copy-selected"
	ReasonOfflineCopy       RouteReason = "offline-copy-selected"
	ReasonPreferredServer   RouteReason = "preferred-goreecloud-server"
	ReasonBestAvailable     RouteReason = "best-authorized-source"
	ReasonTranscodeRequired RouteReason = "transcoding-required"
)

type Candidate struct {
	ProviderID        string
	Kind              domain.SourceKind
	RecordingID       domain.RecordingID
	SourceItemID      domain.SourceItemID
	AssetID           domain.PlayableAssetID
	Match             domain.MatchConfidence
	Availability      domain.AvailabilityState
	Authorized        bool
	RequiresNetwork   bool
	RequiresTranscode bool
}

type Request struct {
	RecordingID       domain.RecordingID
	RequestedProvider string
	NetworkAvailable  bool
	Preference        SourcePreference
	AskWhenMultiple   bool
	AllowNonExact     bool
}

type Decision struct {
	Candidate Candidate
	Reason    RouteReason
}

func Select(req Request, candidates []Candidate) (Decision, error) {
	eligible := make([]Candidate, 0, len(candidates))
	for _, c := range candidates {
		if !c.Authorized || c.RecordingID != req.RecordingID {
			continue
		}
		if c.Match != domain.MatchExact && !req.AllowNonExact {
			continue
		}
		if req.RequestedProvider != "" && c.ProviderID != req.RequestedProvider {
			continue
		}
		if c.Availability == domain.AvailabilityUnavailable {
			continue
		}
		if c.RequiresNetwork && !req.NetworkAvailable && c.Availability == domain.AvailabilityOnline {
			continue
		}
		eligible = append(eligible, c)
	}

	if len(eligible) == 0 {
		return Decision{}, ErrNoPlayableRoute
	}

	if req.RequestedProvider == "" && req.AskWhenMultiple && distinctProviders(eligible) > 1 {
		return Decision{}, ErrUserChoiceRequired
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		si := score(req, eligible[i])
		sj := score(req, eligible[j])
		if si != sj {
			return si > sj
		}
		if eligible[i].ProviderID != eligible[j].ProviderID {
			return eligible[i].ProviderID < eligible[j].ProviderID
		}
		return eligible[i].AssetID < eligible[j].AssetID
	})

	chosen := eligible[0]
	reason := reasonFor(req, chosen)
	if chosen.RequiresTranscode {
		reason = ReasonTranscodeRequired
	}
	return Decision{Candidate: chosen, Reason: reason}, nil
}

func distinctProviders(candidates []Candidate) int {
	seen := make(map[string]struct{})
	for _, c := range candidates {
		seen[c.ProviderID] = struct{}{}
	}
	return len(seen)
}

func score(req Request, c Candidate) int {
	s := 0
	switch c.Match {
	case domain.MatchExact:
		s += 1000
	case domain.MatchEquivalent:
		s += 100
	case domain.MatchAlternate:
		s += 50
	case domain.MatchUnknown:
		s += 10
	}
	switch c.Availability {
	case domain.AvailabilityDownloaded:
		s += 500
	case domain.AvailabilityCached:
		s += 350
	case domain.AvailabilityOffline:
		s += 300
	case domain.AvailabilityOnline:
		s += 100
	}
	if req.Preference == PreferOffline && (c.Availability == domain.AvailabilityDownloaded || c.Availability == domain.AvailabilityCached || c.Availability == domain.AvailabilityOffline) {
		s += 250
	}
	if req.Preference == PreferGoreeCloudServer && c.Kind == domain.SourceGoreeCloudServer {
		s += 250
	}
	if req.RequestedProvider != "" && c.ProviderID == req.RequestedProvider {
		s += 500
	}
	if c.RequiresTranscode {
		s -= 25
	}
	return s
}

func reasonFor(req Request, c Candidate) RouteReason {
	if req.RequestedProvider != "" {
		return ReasonRequestedProvider
	}
	if c.Availability == domain.AvailabilityDownloaded {
		return ReasonDownloadedCopy
	}
	if c.Availability == domain.AvailabilityOffline || c.Availability == domain.AvailabilityCached {
		return ReasonOfflineCopy
	}
	if req.Preference == PreferGoreeCloudServer && c.Kind == domain.SourceGoreeCloudServer {
		return ReasonPreferredServer
	}
	return ReasonBestAvailable
}
