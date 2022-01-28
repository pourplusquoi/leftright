package leftright_test

import (
	"sync"
	"testing"

	"github.com/pourplusquoi/leftright"
	"github.com/stretchr/testify/assert"
)

func TestLeftRight_SingleReader(t *testing.T) {
	reader, writer := leftright.New()

	wg := sync.WaitGroup{}
	wg.Add(2)

	barrier1 := sync.WaitGroup{}
	barrier1.Add(1)
	barrier2 := sync.WaitGroup{}
	barrier2.Add(1)
	barrier3 := sync.WaitGroup{}
	barrier3.Add(1)
	barrier4 := sync.WaitGroup{}
	barrier4.Add(1)
	barrier5 := sync.WaitGroup{}
	barrier5.Add(1)

	go func() {
		defer wg.Done()
		writer.Insert("foo", "hello")
		writer.Insert("bar", "world")

		barrier1.Done()
		barrier2.Wait()

		writer.Publish()
		writer.Insert("baz", "xxx")

		barrier3.Done()
		barrier4.Wait()

		writer.Publish()

		barrier5.Done()
	}()

	go func() {
		defer wg.Done()

		barrier1.Wait()

		_, exists1 := reader.Get("foo")
		_, exists2 := reader.Get("bar")
		assert.Equal(t, false, exists1)
		assert.Equal(t, false, exists2)

		barrier2.Done()
		barrier3.Wait()

		v3, exists3 := reader.Get("foo")
		v4, exists4 := reader.Get("bar")
		_, exists5 := reader.Get("baz")
		assert.Equal(t, true, exists3)
		assert.Equal(t, true, exists4)
		assert.Equal(t, false, exists5)
		assert.Equal(t, "hello", v3)
		assert.Equal(t, "world", v4)

		barrier4.Done()
		barrier5.Wait()

		v6, exists6 := reader.Get("foo")
		v7, exists7 := reader.Get("bar")
		v8, exists8 := reader.Get("baz")
		assert.Equal(t, true, exists6)
		assert.Equal(t, true, exists7)
		assert.Equal(t, true, exists8)
		assert.Equal(t, "hello", v6)
		assert.Equal(t, "world", v7)
		assert.Equal(t, "xxx", v8)

		reader.Drop()
	}()

	wg.Wait()
}

func TestLeftRight_MultiReaders(t *testing.T) {
	n := 8
	reader, writer := leftright.New()
	readers := make([]*leftright.ReadHandle, 0, n)
	readers = append(readers, reader)
	for i := 1; i < n; i++ {
		readers = append(readers, reader.Clone())
	}

	wg := sync.WaitGroup{}
	wg.Add(n + 1)

	barrier1 := sync.WaitGroup{}
	barrier1.Add(1)
	barrier2 := sync.WaitGroup{}
	barrier2.Add(n)
	barrier3 := sync.WaitGroup{}
	barrier3.Add(1)
	barrier4 := sync.WaitGroup{}
	barrier4.Add(n)
	barrier5 := sync.WaitGroup{}
	barrier5.Add(1)

	go func() {
		defer wg.Done()
		writer.Insert("foo", "hello")
		writer.Insert("bar", "world")

		barrier1.Done()
		barrier2.Wait()

		writer.Publish()
		writer.Insert("baz", "xxx")
		writer.Insert("bar", "yyy")
		writer.Remove("foo")

		barrier3.Done()
		barrier4.Wait()

		writer.Publish()

		barrier5.Done()
	}()

	for _, reader := range readers {
		go func(reader *leftright.ReadHandle) {
			defer wg.Done()

			barrier1.Wait()

			_, exists1 := reader.Get("foo")
			_, exists2 := reader.Get("bar")
			assert.Equal(t, false, exists1)
			assert.Equal(t, false, exists2)

			barrier2.Done()
			barrier3.Wait()

			v3, exists3 := reader.Get("foo")
			v4, exists4 := reader.Get("bar")
			_, exists5 := reader.Get("baz")
			assert.Equal(t, true, exists3)
			assert.Equal(t, true, exists4)
			assert.Equal(t, false, exists5)
			assert.Equal(t, "hello", v3)
			assert.Equal(t, "world", v4)

			barrier4.Done()
			barrier5.Wait()

			_, exists6 := reader.Get("foo")
			v7, exists7 := reader.Get("bar")
			v8, exists8 := reader.Get("baz")
			assert.Equal(t, false, exists6)
			assert.Equal(t, true, exists7)
			assert.Equal(t, true, exists8)
			assert.Equal(t, "yyy", v7)
			assert.Equal(t, "xxx", v8)

			reader.Drop()
		}(reader)
	}

	wg.Wait()
}

func TestLeftRight_NoDataRace(t *testing.T) {
	n, m := 1024, 1000
	reader, writer := leftright.New()
	readers := make([]*leftright.ReadHandle, 0, n)
	readers = append(readers, reader)
	for i := 1; i < n; i++ {
		readers = append(readers, reader.Clone())
	}

	wg := sync.WaitGroup{}
	wg.Add(n + 1)

	go func() {
		defer wg.Done()
		for i := 0; i < m; i++ {
			writer.Insert(i, i)
		}
		writer.Publish()
		for i := 0; i < m; i++ {
			writer.Insert(i, i*2)
		}
		writer.Publish()
		for i := 0; i < m; i++ {
			writer.Remove(i)
		}
		writer.Publish()
	}()

	for _, reader := range readers {
		go func(reader *leftright.ReadHandle) {
			defer wg.Done()
			for i := 0; i < m; i++ {
				_, _ = reader.Get(i)
			}
			reader.Drop()
		}(reader)
	}

	wg.Wait()

	assert.Equal(t, 0, reader.Len())
}
