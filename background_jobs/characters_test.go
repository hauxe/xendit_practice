package jobs

import (
	"encoding/json"
	"testing"

	"github.com/hauxe/xendit_pratice/cacher"
	"github.com/hauxe/xendit_pratice/marvel"
	"github.com/hauxe/xendit_pratice/test"
	"github.com/stretchr/testify/require"
)

func TestUpdateMarvelCharacterList(t *testing.T) {
	t.Parallel()
	t.Run("no_update", func(t *testing.T) {
		t.Parallel()
		c := cacher.NewCacher()
		cacheKey := "test_get_all_character"
		val := "[0,1,2]"
		c.Set(cacheKey, val)
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData1stCall))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		updateMarvelCharacterList(c, cacheKey, api)

		list, ok := c.Get(cacheKey)
		require.True(t, ok)
		require.Equal(t, val, list)
	})
	t.Run("updated", func(t *testing.T) {
		t.Parallel()
		c := cacher.NewCacher()
		cacheKey := "test_get_all_character"
		val := "[0,1,2]"
		c.Set(cacheKey, val)
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		updateMarvelCharacterList(c, cacheKey, api)

		v, ok := c.Get(cacheKey)
		require.True(t, ok)
		var list []int
		err = json.Unmarshal([]byte(v), &list)
		require.NoError(t, err)
		require.Len(t, list, 1000)
	})
}

func TestCheckMarvelUpdate(t *testing.T) {
	t.Parallel()
	t.Run("cache_not_found", func(t *testing.T) {
		t.Parallel()
		c := cacher.NewCacher()
		cacheKey := "test_get_all_character"
		shouldUpdate, err := checkMarvelUpdate(c, cacheKey, nil)
		require.NoError(t, err)
		require.True(t, shouldUpdate)
	})
	t.Run("cache_invalid_json", func(t *testing.T) {
		t.Parallel()
		c := cacher.NewCacher()
		cacheKey := "test_get_all_character"
		c.Set(cacheKey, "invalid json")
		shouldUpdate, err := checkMarvelUpdate(c, cacheKey, nil)
		require.NoError(t, err)
		require.True(t, shouldUpdate)
	})
	t.Run("get_total_count_error", func(t *testing.T) {
		t.Parallel()
		c := cacher.NewCacher()
		cacheKey := "test_get_all_character"
		c.Set(cacheKey, "[0,1,2]")
		testServer, err := test.NewTestServer(test.NewMockHandler("invalid json"))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		shouldUpdate, err := checkMarvelUpdate(c, cacheKey, api)
		require.Error(t, err)
		require.False(t, shouldUpdate)
	})
	t.Run("list_unchanged", func(t *testing.T) {
		t.Parallel()
		c := cacher.NewCacher()
		cacheKey := "test_get_all_character"
		c.Set(cacheKey, "[0,1,2]")
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData1stCall))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		shouldUpdate, err := checkMarvelUpdate(c, cacheKey, api)
		require.NoError(t, err)
		require.False(t, shouldUpdate)
	})
	t.Run("list_changed", func(t *testing.T) {
		t.Parallel()
		c := cacher.NewCacher()
		cacheKey := "test_get_all_character"
		c.Set(cacheKey, "[0,1,2]")
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		shouldUpdate, err := checkMarvelUpdate(c, cacheKey, api)
		require.NoError(t, err)
		require.True(t, shouldUpdate)
	})
}
