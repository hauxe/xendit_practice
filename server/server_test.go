package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hauxe/xendit_pratice/cacher"
	"github.com/hauxe/xendit_pratice/marvel"
	"github.com/hauxe/xendit_pratice/test"
	"github.com/stretchr/testify/require"
)

func TestGetListCharacters(t *testing.T) {
	t.Parallel()
	t.Run("cached", func(t *testing.T) {
		t.Parallel()
		s := &Server{
			marvelAPI: marvel.NewAPI("", "", ""),
			cacher:    cacher.NewCacher(),
		}
		s.cacher.Set(Characters_Cache_Key, "[0,1,2]")
		rec := httptest.NewRecorder()
		s.GetListCharacters(rec, nil)
		require.NotNil(t, rec.Body)
		dec := json.NewDecoder(rec.Body)
		var list []int
		err := dec.Decode(&list)
		require.NoError(t, err)
		require.EqualValues(t, []int{0, 1, 2}, list)
	})
	t.Run("uncached", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData1stCall))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		s := &Server{
			marvelAPI: api,
			cacher:    cacher.NewCacher(),
		}
		rec := httptest.NewRecorder()
		s.GetListCharacters(rec, nil)
		require.NotNil(t, rec.Body)
		dec := json.NewDecoder(rec.Body)
		var list []int
		err = dec.Decode(&list)
		require.NoError(t, err)
		require.EqualValues(t, []int{1011334, 1011335, 1011336}, list)
	})
	t.Run("stress", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData1stCall))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		s := &Server{
			marvelAPI: api,
			cacher:    cacher.NewCacher(),
		}
		n := 1000
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				rec := httptest.NewRecorder()
				s.GetListCharacters(rec, nil)
				require.NotNil(t, rec.Body)
				dec := json.NewDecoder(rec.Body)
				var list []int
				err = dec.Decode(&list)
				require.NoError(t, err)
				require.EqualValues(t, []int{1011334, 1011335, 1011336}, list)
			}()
		}
		wg.Wait()
	})
}

func TestGetCharacterInfo(t *testing.T) {
	t.Parallel()
	t.Run("cached", func(t *testing.T) {
		t.Parallel()
		id := 11111
		info := &marvel.MarvelCharacter{
			ID:          id,
			Name:        "test",
			Description: "test description",
		}
		b, err := json.Marshal(info)
		require.NoError(t, err)
		cacheKey := buildCharacterInfoCacheKey(id)
		s := &Server{
			marvelAPI: marvel.NewAPI("", "", ""),
			cacher:    cacher.NewCacher(),
		}
		s.cacher.Set(cacheKey, string(b))
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/v1/characters/", nil)
		require.NoError(t, err)
		vars := map[string]string{
			"id": strconv.Itoa(id),
		}
		r := mux.SetURLVars(req, vars)

		s.GetCharacterInfo(rec, r)
		require.NotNil(t, rec.Body)
		dec := json.NewDecoder(rec.Body)
		var result marvel.MarvelCharacter
		err = dec.Decode(&result)
		require.NoError(t, err)
		require.Equal(t, info.ID, result.ID)
		require.Equal(t, info.Name, result.Name)
		require.Equal(t, info.Description, result.Description)
	})
	t.Run("uncached", func(t *testing.T) {
		t.Parallel()
		id := 1011334
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData1stCall))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		s := &Server{
			marvelAPI: api,
			cacher:    cacher.NewCacher(),
		}
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/v1/characters/", nil)
		require.NoError(t, err)
		vars := map[string]string{
			"id": strconv.Itoa(id),
		}
		r := mux.SetURLVars(req, vars)
		s.GetCharacterInfo(rec, r)
		require.NotNil(t, rec.Body)
		dec := json.NewDecoder(rec.Body)
		var result marvel.MarvelCharacter
		err = dec.Decode(&result)
		require.NoError(t, err)
		require.Equal(t, id, result.ID)
		require.Equal(t, "1-D Man", result.Name)
		require.Equal(t, "test description", result.Description)
	})
	t.Run("stress", func(t *testing.T) {
		t.Parallel()
		id := 1011334
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData1stCall))
		require.NoError(t, err)
		api := marvel.NewAPI(testServer.URL, "", "")
		s := &Server{
			marvelAPI: api,
			cacher:    cacher.NewCacher(),
		}
		req, err := http.NewRequest(http.MethodGet, "/v1/characters/", nil)
		require.NoError(t, err)
		vars := map[string]string{
			"id": strconv.Itoa(id),
		}
		r := mux.SetURLVars(req, vars)
		n := 1000
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				rec := httptest.NewRecorder()
				s.GetCharacterInfo(rec, r)
				require.NotNil(t, rec.Body)
				dec := json.NewDecoder(rec.Body)
				var result marvel.MarvelCharacter
				err = dec.Decode(&result)
				require.NoError(t, err)
				require.Equal(t, id, result.ID)
				require.Equal(t, "1-D Man", result.Name)
				require.Equal(t, "test description", result.Description)
			}()
		}
		wg.Wait()
	})
}
