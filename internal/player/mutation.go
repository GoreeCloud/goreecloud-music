package player

import (
	"errors"
	"strings"
)

var ErrInvalidTrackID = errors.New("invalid track id")

func (q *Queue) Append(trackIDs ...string) error {
	if len(trackIDs) == 0 {
		return nil
	}
	for _, trackID := range trackIDs {
		if strings.TrimSpace(trackID) == "" {
			return ErrInvalidTrackID
		}
	}

	wasEmpty := len(q.TrackIDs) == 0
	q.TrackIDs = append(q.TrackIDs, trackIDs...)
	if wasEmpty {
		q.Current = 0
	}
	return nil
}

func (q *Queue) Remove(index int) (string, error) {
	if index < 0 || index >= len(q.TrackIDs) {
		return "", ErrInvalidQueueIndex
	}

	removed := q.TrackIDs[index]
	q.TrackIDs = append(q.TrackIDs[:index], q.TrackIDs[index+1:]...)
	if len(q.TrackIDs) == 0 {
		q.Current = -1
		return removed, nil
	}
	if index < q.Current {
		q.Current--
	} else if index == q.Current && q.Current >= len(q.TrackIDs) {
		q.Current = len(q.TrackIDs) - 1
	}
	return removed, nil
}
