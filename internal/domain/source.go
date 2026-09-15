package domain

type SourceKind string

const (
	SourceGoreeCloudServer SourceKind = "goreecloud-server"
	SourceExternal         SourceKind = "external-provider"
	SourceInternetRadio    SourceKind = "internet-radio"
)

type MatchConfidence string

const (
	MatchExact      MatchConfidence = "exact"
	MatchEquivalent MatchConfidence = "equivalent"
	MatchAlternate  MatchConfidence = "alternate"
	MatchUnknown    MatchConfidence = "unknown"
)

type AvailabilityState string

const (
	AvailabilityOnline      AvailabilityState = "online"
	AvailabilityOffline     AvailabilityState = "offline"
	AvailabilityDownloaded  AvailabilityState = "downloaded"
	AvailabilityCached      AvailabilityState = "cached"
	AvailabilityUnavailable AvailabilityState = "unavailable"
)

type SourceItem struct {
	ID          SourceItemID
	ProviderID  string
	Kind        SourceKind
	RecordingID RecordingID
	ReleaseID   ReleaseID
}

type PlayableAsset struct {
	ID              PlayableAssetID
	SourceItemID    SourceItemID
	Codec           string
	BitrateKbps     int
	BitDepth        int
	SampleRateHz    int
	Channels        int
	RequiresNetwork bool
}

type AvailabilityInput struct {
	AuthorizationValid bool
	LocalReadable      bool
	IntegrityVerified  bool
	Downloaded         bool
	Cached             bool
	NetworkRequired    bool
	SourceReachable    bool
	Expired            bool
	Corrupt            bool
}

func DeriveAvailability(in AvailabilityInput) AvailabilityState {
	if !in.AuthorizationValid || in.Expired || in.Corrupt {
		return AvailabilityUnavailable
	}
	if in.Downloaded && in.LocalReadable && in.IntegrityVerified {
		return AvailabilityDownloaded
	}
	if in.Cached && in.LocalReadable && in.IntegrityVerified {
		return AvailabilityCached
	}
	if in.LocalReadable && !in.NetworkRequired {
		return AvailabilityOffline
	}
	if in.SourceReachable {
		return AvailabilityOnline
	}
	return AvailabilityUnavailable
}
