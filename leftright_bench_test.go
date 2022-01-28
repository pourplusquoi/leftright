package leftright_test

import (
	"sync"
	"testing"

	"github.com/pourplusquoi/leftright"
)

const n = 4

func BenchmarkLeftRight(b *testing.B) {
	reader, writer := leftright.New()
	readers := make([]*leftright.ReadHandle, 0, n)
	readers = append(readers, reader)
	for i := 1; i < n; i++ {
		readers = append(readers, reader.Clone())
	}

	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}
		wg.Add(n + 1)

		barrier1 := sync.WaitGroup{}
		barrier1.Add(n + 1)
		barrier2 := sync.WaitGroup{}
		barrier2.Add(1)

		go func() {
			defer wg.Done()
			barrier1.Done()
			barrier2.Wait()
			for i := 0; i < 2000; i++ {
				writer.Insert(i, i)
				writer.Insert(i+1, i)
				writer.Insert(i+2, i)
				writer.Insert(i+3, i)
				writer.Insert(i+4, i)
				writer.Remove(i)
				writer.Remove(i + 2)
				writer.Remove(i + 4)
			}
			writer.Publish()
		}()

		for _, reader := range readers {
			go func(reader *leftright.ReadHandle) {
				defer wg.Done()
				barrier1.Done()
				barrier2.Wait()
				for i := 0; i < 2000; i++ {
					_, _ = reader.Get(i)
				}
			}(reader)
		}

		barrier1.Wait()
		barrier2.Done()
		wg.Wait()
	}
}

func BenchmarkSyncMap(b *testing.B) {
	syncmap := sync.Map{}

	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}
		wg.Add(n + 1)

		barrier1 := sync.WaitGroup{}
		barrier1.Add(n + 1)
		barrier2 := sync.WaitGroup{}
		barrier2.Add(1)

		go func() {
			defer wg.Done()
			barrier1.Done()
			barrier2.Wait()
			for i := 0; i < 2000; i++ {
				syncmap.Store(i, i)
				syncmap.Store(i+1, i)
				syncmap.Store(i+2, i)
				syncmap.Store(i+3, i)
				syncmap.Store(i+4, i)
				syncmap.Delete(i)
				syncmap.Delete(i + 2)
				syncmap.Delete(i + 4)
			}
		}()

		for j := 0; j < n; j++ {
			go func() {
				defer wg.Done()
				barrier1.Done()
				barrier2.Wait()
				for i := 0; i < 2000; i++ {
					_, _ = syncmap.Load(i)
				}
			}()
		}

		barrier1.Wait()
		barrier2.Done()
		wg.Wait()
	}
}
