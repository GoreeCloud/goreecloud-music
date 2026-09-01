package player

import (
	"errors"
	"math/rand"
)

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

// Shuffle randomizes only the unplayed portion of the queue. The current track
// and already-played prefix stay in place so enabling shuffle does not rewrite
// listening history or interrupt the active track.
func (q *Queue) Shuffle(source rand.Source) {
	if source == nil || len(q.TrackIDs) < 2 {
		return
	}

	start := 0
	if q.Current >= 0 && q.Current < len(q.TrackIDs) {
		start = q.Current + 1
	}
	if len(q.TrackIDs)-start < 2 {
		return
	}

	rng := rand.New(source)
	rng.Shuffle(len(q.TrackIDs)-start, func(i, j int) {
		q.TrackIDs[start+i], q.TrackIDs[start+j] = q.TrackIDs[start+j], q.TrackIDs[start+i]
	})
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
