package leftright

import (
	"sync"

	"github.com/pourplusquoi/slabmap"
)

// New creates a pair of `ReadHandle` and `WriteHandle`, based on a concurrency
// model called "left-right".
//
// Basically, there are two copies of the same map in the concurrency model,
// where the writer writes to the one map and readers read from the other map.
// Only after the writer explicitly publishes changes can readers see the
// modifications until then.
//
// The benefit of "left-right" concurrency model is that reads are entirely
// lock-free and can be very fast, and that the writer can decide when to
// publish changes to readers. However, the cost is that the model consumes
// extra memory, and that the model only supports single writer, and that the
// writer has to do extra work while writing.
//
// To summarize, this concurrency model is suitable for use cases where reads
// are more frequent than writes and strong consistency is not required.
func New() (*ReadHandle, *WriteHandle) {
	slab := newSyncSlabMap()
	epoch := uint64(0)
	index := slab.inner.Insert(&epoch)
	parent := &leftrightMap{
		maps: [2]map[interface{}]interface{}{
			make(map[interface{}]interface{}),
			make(map[interface{}]interface{}),
		},
		current: 0,
		epochs:  slab,
	}
	return newReadHandle(parent, &epoch, index), newWriteHandle(parent)
}

type leftrightMap struct {
	maps    [2]map[interface{}]interface{}
	current uint32
	epochs  *syncSlabMap
}

type syncSlabMap struct {
	inner *slabmap.SlabMap
	mu    sync.Mutex
}

func newSyncSlabMap() *syncSlabMap {
	return &syncSlabMap{
		inner: slabmap.NewSlabMap(),
	}
}

func (s *syncSlabMap) Lock() *slabmap.SlabMap {
	s.mu.Lock()
	return s.inner
}

func (s *syncSlabMap) Unlock() {
	s.mu.Unlock()
}
