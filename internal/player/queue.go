package player

import "errors"

var ErrInvalidQueueIndex = errors.New("invalid queue index")

type Queue struct {
	TrackIDs []string
	Current  int
	Repeat   RepeatMode
}

type RepeatMode string

const (
	RepeatOff RepeatMode = "off"
	RepeatAll RepeatMode = "all"
	RepeatOne RepeatMode = "one"
)

func NewQueue(trackIDs []string) Queue {
	copyIDs := append([]string(nil), trackIDs...)
	current := -1
	if len(copyIDs) > 0 {
		current = 0
	}
	return Queue{TrackIDs: copyIDs, Current: current, Repeat: RepeatOff}
}

func (q Queue) CurrentTrackID() (string, bool) {
	if q.Current < 0 || q.Current >= len(q.TrackIDs) {
		return "", false
	}
	return q.TrackIDs[q.Current], true
}

func (q *Queue) SetCurrent(index int) error {
	if index < 0 || index >= len(q.TrackIDs) {
		return ErrInvalidQueueIndex
	}
	q.Current = index
	return nil
}

func (q *Queue) Next() bool {
	if len(q.TrackIDs) == 0 {
		q.Current = -1
		return false
	}
	if q.Repeat == RepeatOne && q.Current >= 0 {
		return true
	}
	if q.Current+1 < len(q.TrackIDs) {
		q.Current++
		return true
	}
	if q.Repeat == RepeatAll {
		q.Current = 0
		return true
	}
	return false
}

func (q *Queue) Previous() bool {
	if len(q.TrackIDs) == 0 {
		q.Current = -1
		return false
	}
	if q.Repeat == RepeatOne && q.Current >= 0 {
		return true
	}
	if q.Current > 0 {
		q.Current--
		return true
	}
	if q.Repeat == RepeatAll {
		q.Current = len(q.TrackIDs) - 1
		return true
	}
	return false
}
