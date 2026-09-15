package authorization

import (
	"context"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

type Action string

const (
	ActionSearch   Action = "search"
	ActionRead     Action = "read"
	ActionPlay     Action = "play"
	ActionDownload Action = "download"
	ActionShare    Action = "share"
	ActionAdmin    Action = "admin"
)

type Request struct {
	ProfileID domain.ProfileID
	Action    Action
	Resource  string
	Purpose   string
}

type Decision struct {
	Allowed   bool
	Reason    string
	ExpiresAt time.Time
}

type Authorizer interface {
	Authorize(context.Context, Request) (Decision, error)
}
