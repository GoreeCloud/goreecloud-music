package player

import (
	"bytes"
	"errors"
)

const MaxQueueRecordPayloadBytes = 64 * 1024

var ErrInvalidQueueRecord = errors.New("invalid queue record")

// QueueRecord is a storage-neutral persistence value. The encoded queue payload
// is kept private so callers cannot mutate a record without passing through the
// versioned snapshot and optimistic-revision boundaries.
type QueueRecord struct {
	revision QueueRevision
	payload  []byte
}

func NewQueueRecord(queue Queue) (QueueRecord, error) {
	payload, err := EncodeQueueSnapshot(queue)
	if err != nil {
		return QueueRecord{}, err
	}
	return queueRecordFromCanonicalPayload(0, payload)
}

// RestoreQueueRecord reconstructs a record returned by an external persistence
// adapter. Only bounded canonical queue payloads are accepted.
func RestoreQueueRecord(revision QueueRevision, payload []byte) (QueueRecord, error) {
	if len(payload) == 0 || len(payload) > MaxQueueRecordPayloadBytes {
		return QueueRecord{}, ErrInvalidQueueRecord
	}
	queue, err := DecodeQueueSnapshot(payload)
	if err != nil {
		return QueueRecord{}, err
	}
	canonical, err := EncodeQueueSnapshot(queue)
	if err != nil || !bytes.Equal(payload, canonical) {
		return QueueRecord{}, ErrInvalidQueueRecord
	}
	return queueRecordFromCanonicalPayload(revision, canonical)
}

func queueRecordFromCanonicalPayload(revision QueueRevision, payload []byte) (QueueRecord, error) {
	if len(payload) == 0 || len(payload) > MaxQueueRecordPayloadBytes {
		return QueueRecord{}, ErrInvalidQueueRecord
	}
	return QueueRecord{revision: revision, payload: append([]byte(nil), payload...)}, nil
}

func (r QueueRecord) Revision() QueueRevision {
	return r.revision
}

func (r QueueRecord) Payload() []byte {
	return append([]byte(nil), r.payload...)
}

func (r QueueRecord) Restore() (Queue, error) {
	return DecodeQueueSnapshot(r.payload)
}

// Update applies a queue replacement only when expected matches the record's
// current revision. Failed validation or a stale writer leaves the original
// record unchanged.
func (r QueueRecord) Update(expected QueueRevision, queue Queue) (QueueRecord, error) {
	nextRevision, err := AdvanceQueueRevision(expected, r.revision)
	if err != nil {
		return r, err
	}
	payload, err := EncodeQueueSnapshot(queue)
	if err != nil {
		return r, err
	}
	next, err := queueRecordFromCanonicalPayload(nextRevision, payload)
	if err != nil {
		return r, err
	}
	return next, nil
}
