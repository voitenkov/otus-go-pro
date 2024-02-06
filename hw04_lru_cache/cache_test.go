package hw04lrucache

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic: size exceeded", func(t *testing.T) {
		c := NewCache(3)

		for i := 1; i < 5; i++ {
			wasInCache := c.Set(Key(strconv.Itoa(i)), i)
			require.False(t, wasInCache)
		}

		val, ok := c.Get("1")
		require.False(t, ok)
		require.Nil(t, val)

		for i := 2; i < 5; i++ {
			val, ok := c.Get(Key(strconv.Itoa(i)))
			require.True(t, ok)
			require.Equal(t, i, val)
		}
	})

	t.Run("purge logic: long used", func(t *testing.T) {
		c := NewCache(3)

		for i := 1; i < 4; i++ {
			wasInCache := c.Set(Key(strconv.Itoa(i)), i)
			require.False(t, wasInCache)
		}

		for i := 1; i < 4; i++ {
			val, ok := c.Get(Key(strconv.Itoa(i)))
			require.True(t, ok)
			require.Equal(t, i, val)
		}

		wasInCache := c.Set("4", 4)
		require.False(t, wasInCache)

		val, ok := c.Get("1")
		require.False(t, ok)
		require.Nil(t, val)
	})
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
	require.Equal(t, 2, runtime.NumGoroutine())
}

func TestCacheMultithreadingWithLenAndClear(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			go func(i int) {
				defer wg.Done()
				c.Set(Key(strconv.Itoa(i)), i)
			}(i)
		} else {
			go func() {
				defer wg.Done()
				c.Get(Key(strconv.Itoa(rand.Intn(9))))
			}()
		}
	}
	wg.Wait()
	require.Equal(t, 5, c.Len())
	c.Clear()
	require.Equal(t, 0, c.Len())
}
