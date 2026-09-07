package player

import "errors"

var ErrInvalidQueueSnapshot = errors.New("invalid queue snapshot")

type QueueSnapshot struct {
	TrackIDs []string
	Current  int
	Repeat   RepeatMode
}

func (q Queue) Snapshot() QueueSnapshot {
	return QueueSnapshot{TrackIDs: append([]string(nil), q.TrackIDs...), Current: q.Current, Repeat: q.Repeat}
}

func RestoreQueue(snapshot QueueSnapshot) (Queue, error) {
	if !validRepeatMode(snapshot.Repeat) {
		return Queue{}, ErrInvalidQueueSnapshot
	}
	if len(snapshot.TrackIDs) == 0 {
		if snapshot.Current != -1 {
			return Queue{}, ErrInvalidQueueSnapshot
		}
	} else if snapshot.Current < 0 || snapshot.Current >= len(snapshot.TrackIDs) {
		return Queue{}, ErrInvalidQueueSnapshot
	}
	return Queue{TrackIDs: append([]string(nil), snapshot.TrackIDs...), Current: snapshot.Current, Repeat: snapshot.Repeat}, nil
}

func validRepeatMode(mode RepeatMode) bool {
	return mode == RepeatOff || mode == RepeatAll || mode == RepeatOne
}
