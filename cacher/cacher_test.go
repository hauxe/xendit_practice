package cacher

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Parallel()
	t.Run("cache_hit", func(t *testing.T) {
		t.Parallel()
		c := &cache{
			storage: make(map[string]string),
		}
		cacheKey := "test_cache_hit"
		cacheValue := "hit"
		c.storage[cacheKey] = cacheValue
		v, ok := c.Get(cacheKey)
		require.True(t, ok)
		require.Equal(t, cacheValue, v)
	})
	t.Run("cache_miss", func(t *testing.T) {
		t.Parallel()
		c := &cache{
			storage: make(map[string]string),
		}
		cacheKey := "test_cache_miss"
		v, ok := c.Get(cacheKey)
		require.False(t, ok)
		require.Empty(t, v)
	})
}

func TestSet(t *testing.T) {
	t.Parallel()
	t.Run("new", func(t *testing.T) {
		t.Parallel()
		c := &cache{
			storage: make(map[string]string),
		}
		cacheKey := "test_cache_set_new"
		cacheValue := "new"
		c.Set(cacheKey, cacheValue)
		v, ok := c.Get(cacheKey)
		require.True(t, ok)
		require.Equal(t, cacheValue, v)
	})
	t.Run("overwrite", func(t *testing.T) {
		t.Parallel()
		c := &cache{
			storage: make(map[string]string),
		}
		cacheKey := "test_cache_set_overwrite"
		cacheValue := "overwrite"
		c.storage[cacheKey] = "new"
		c.Set(cacheKey, cacheValue)
		v, ok := c.Get(cacheKey)
		require.True(t, ok)
		require.Equal(t, cacheValue, v)
	})
}

func TestStress(t *testing.T) {
	t.Parallel()
	// Stress Test cache used for race detection
	n := 1000
	c := NewCacher()
	var wg sync.WaitGroup
	wg.Add(2 * n)
	for i := 0; i < n; i++ {
		k := strconv.Itoa(i % 13)
		cacheKey := "test_stress_" + k
		go func(i int) {
			defer wg.Done()
			c.Set(cacheKey, k)
		}(i)
		go func(i int) {
			defer wg.Done()
			c.Get(cacheKey)
		}(i)
	}
	wg.Wait()
}
