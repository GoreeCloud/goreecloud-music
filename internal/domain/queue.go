package domain

import "fmt"

type QueueItem struct {
	ID                QueueItemID
	OwnerProfileID    ProfileID
	RecordingID       RecordingID
	SourceItemID       SourceItemID
	RequestedProvider string
	SelectedAssetID   PlayableAssetID
	RouteReason       string
	Position          int
	Revision          uint64
}

func (q QueueItem) Validate() error {
	if q.ID == "" || q.OwnerProfileID == "" || q.RecordingID == "" || q.SourceItemID == "" {
		return fmt.Errorf("queue item requires stable queue, owner, recording, and source identities")
	}
	if q.Position < 0 {
		return fmt.Errorf("queue position must be non-negative")
	}
	return nil
}
