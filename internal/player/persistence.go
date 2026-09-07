package player

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

const QueueSnapshotSchemaVersion = 1

var ErrUnsupportedQueueSnapshotVersion = errors.New("unsupported queue snapshot version")

type persistedQueueSnapshot struct {
	SchemaVersion int        `json:"schemaVersion"`
	TrackIDs      []string   `json:"trackIds"`
	Current       int        `json:"current"`
	Repeat        RepeatMode `json:"repeat"`
}

// EncodeQueueSnapshot creates a storage-neutral, versioned queue payload. It
// does not choose a database, filesystem, user identity, or synchronization
// mechanism; those remain integration responsibilities outside the queue core.
func EncodeQueueSnapshot(q Queue) ([]byte, error) {
	snapshot := q.Snapshot()
	if _, err := RestoreQueue(snapshot); err != nil {
		return nil, err
	}
	return json.Marshal(persistedQueueSnapshot{
		SchemaVersion: QueueSnapshotSchemaVersion,
		TrackIDs:      snapshot.TrackIDs,
		Current:       snapshot.Current,
		Repeat:        snapshot.Repeat,
	})
}

func DecodeQueueSnapshot(data []byte) (Queue, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var persisted persistedQueueSnapshot
	if err := decoder.Decode(&persisted); err != nil {
		return Queue{}, ErrInvalidQueueSnapshot
	}
	if persisted.SchemaVersion != QueueSnapshotSchemaVersion {
		return Queue{}, ErrUnsupportedQueueSnapshotVersion
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Queue{}, ErrInvalidQueueSnapshot
	}

	return RestoreQueue(QueueSnapshot{
		TrackIDs: persisted.TrackIDs,
		Current:  persisted.Current,
		Repeat:   persisted.Repeat,
	})
}
