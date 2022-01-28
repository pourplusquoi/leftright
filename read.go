package leftright

import (
	"sync/atomic"
)

// ReadHandle cannot be shared between goroutines safely. Prefer calling
// `Clone` to create a new `ReadHandle` for every new goroutine.
type ReadHandle struct {
	parent *leftrightMap
	epoch  *uint64
	index  int
}

// Len is thread safe.
func (h *ReadHandle) Len() int {
	atomic.AddUint64(h.epoch, 1)
	length := len(h.currentReadMap())
	atomic.AddUint64(h.epoch, 1)
	return length
}

// Get is thread safe.
func (h *ReadHandle) Get(key interface{}) (value interface{}, exists bool) {
	atomic.AddUint64(h.epoch, 1)
	value, exists = h.currentReadMap()[key]
	atomic.AddUint64(h.epoch, 1)
	return value, exists
}

// Clone is thread safe.
func (h *ReadHandle) Clone() *ReadHandle {
	epoch := uint64(0)
	index := h.parent.epochs.Lock().Insert(&epoch)
	h.parent.epochs.Unlock()
	return newReadHandle(h.parent, &epoch, index)
}

// Drop is thread safe.
func (h *ReadHandle) Drop() {
	_, _ = h.parent.epochs.Lock().Remove(h.index)
	h.parent.epochs.Unlock()
}

func (h *ReadHandle) currentReadMap() map[interface{}]interface{} {
	current := atomic.LoadUint32(&h.parent.current)
	return h.parent.maps[current%2]
}

func newReadHandle(parent *leftrightMap, epoch *uint64, index int) *ReadHandle {
	return &ReadHandle{
		parent: parent,
		epoch:  epoch,
		index:  index,
	}
}
