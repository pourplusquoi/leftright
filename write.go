package leftright

import (
	"runtime"
	"sync/atomic"
)

// WriteHandle cannot be shared between goroutines safely. In fact, the
// "left-right" concurrency model only permits one single writer.
type WriteHandle struct {
	parent *leftrightMap
	oplogs []oplog
	swap   uint64
}

// Insert is not thread safe.
func (h *WriteHandle) Insert(key, value interface{}) {
	h.oplogs = append(h.oplogs, newInsert(key, value))
	h.currentWriteMap()[key] = value
}

// Remove is not thread safe.
func (h *WriteHandle) Remove(key interface{}) {
	h.oplogs = append(h.oplogs, newRemove(key))
	delete(h.currentWriteMap(), key)
}

// Publish is not thread safe.
func (h *WriteHandle) Publish() {
	atomic.AddUint32(&h.parent.current, 1)
	h.waitForRead()
	h.applyOplog()
}

func (h *WriteHandle) applyOplog() {
	writeMap := h.currentWriteMap()
	for _, record := range h.oplogs {
		switch record.op {
		case operationInsert:
			writeMap[record.key] = record.value
		case operationRemove:
			delete(writeMap, record.key)
		default:
			// Do nothing.
		}
	}
	h.oplogs = make([]oplog, 0)
}

func (h *WriteHandle) waitForRead() {
	type cached struct {
		was uint64
		ptr *uint64
	}
	odds := make(map[int]cached)

	h.parent.epochs.Lock().Range(func(key int, value interface{}) bool {
		ptr := value.(*uint64)
		epoch := atomic.LoadUint64(ptr)
		if epoch%2 == 1 {
			odds[key] = cached{
				was: epoch,
				ptr: ptr,
			}
		}
		return true
	})
	h.parent.epochs.Unlock()

	for len(odds) > 0 {
		runtime.Gosched()
		for key, cached := range odds {
			if atomic.LoadUint64(cached.ptr) != cached.was {
				delete(odds, key)
			}
		}
	}
}

func (h *WriteHandle) currentWriteMap() map[interface{}]interface{} {
	current := (atomic.LoadUint32(&h.parent.current) + 1) % 2
	return h.parent.maps[current]
}

func newWriteHandle(parent *leftrightMap) *WriteHandle {
	return &WriteHandle{
		parent: parent,
		oplogs: make([]oplog, 0),
	}
}
