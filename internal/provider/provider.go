package provider

import (
	"context"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

type Capabilities struct {
	Search           bool
	Browse           bool
	Playback         bool
	Seek             bool
	RangeRequests    bool
	Artwork          bool
	Lyrics           bool
	OfflineRetention bool
	QualityProfiles  []string
}

type Health struct {
	Available  bool
	Degraded   bool
	Reason     string
	RetryAfter time.Duration
}

type SearchRequest struct {
	ProfileID domain.ProfileID
	Query     string
	Limit     int
}

type SearchResult struct {
	SourceItem domain.SourceItem
	Title      string
	Artist     string
	Album      string
	DurationMs int64
	Match      domain.MatchConfidence
}

type PlaybackRequest struct {
	ProfileID    domain.ProfileID
	SourceItemID domain.SourceItemID
	Quality      string
}

type PlaybackResult struct {
	Asset     domain.PlayableAsset
	ExpiresAt time.Time
}

type Adapter interface {
	ID() string
	Version() string
	Capabilities(context.Context) (Capabilities, error)
	Health(context.Context) (Health, error)
	Search(context.Context, SearchRequest) ([]SearchResult, error)
	ResolvePlayback(context.Context, PlaybackRequest) (PlaybackResult, error)
}
