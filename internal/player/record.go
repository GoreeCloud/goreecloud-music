package player

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
	return QueueRecord{revision: 0, payload: append([]byte(nil), payload...)}, nil
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
	return QueueRecord{
		revision: nextRevision,
		payload:  append([]byte(nil), payload...),
	}, nil
}
