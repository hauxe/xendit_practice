package marvel

import (
	"context"
	"testing"

	"github.com/hauxe/xendit_pratice/test"
	"github.com/stretchr/testify/require"
)

func TestGetCharacterInfo(t *testing.T) {
	t.Parallel()
	t.Run("host_error", func(t *testing.T) {
		t.Parallel()
		api := &API{
			host: "test_failed_host",
		}
		info, err := api.GetCharacterInfo(0)
		require.Error(t, err)
		require.Nil(t, info)
	})
	t.Run("api_result_json_error", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler("invalid json"))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		info, err := api.GetCharacterInfo(0)
		require.Error(t, err)
		require.Nil(t, info)
	})
	t.Run("api_result_invalid_code", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(`{"code": 409}`))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		info, err := api.GetCharacterInfo(0)
		require.Error(t, err)
		require.Nil(t, info)
	})
	t.Run("api_result_invalid_data", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(`{"code": 200}`))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		info, err := api.GetCharacterInfo(0)
		require.Error(t, err)
		require.Nil(t, info)
	})
	t.Run("api_result_mismatch_id", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleJsonFromMarvel))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		info, err := api.GetCharacterInfo(0)
		require.Error(t, err)
		require.Nil(t, info)
	})
	t.Run("api_result_success", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleJsonFromMarvel))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		id := 1011334
		info, err := api.GetCharacterInfo(id)
		require.NoError(t, err)
		require.Equal(t, id, info.ID)
		require.Equal(t, "3-D Man", info.Name)
		require.Equal(t, "test description", info.Description)
	})
}

func TestGetAllCharacters(t *testing.T) {
	t.Parallel()
	t.Run("stop_at_first_call", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData1stCall))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		list, err := api.GetAllCharacters()
		require.NoError(t, err)
		require.Len(t, list, 3)
		require.EqualValues(t, []int{1011334, 1011335, 1011336}, list)
	})
	t.Run("large_list", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleAllData))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		list, err := api.GetAllCharacters()
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(list), 1000)
	})
	t.Run("flaky_error", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockFlakyHandler(test.SampleAllData, 100))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		list, err := api.GetAllCharacters()
		require.Error(t, err)
		require.Empty(t, list)
	})
}

func TestGetCharacterListJob(t *testing.T) {
	t.Parallel()
	t.Run("test_get_index_fail", func(t *testing.T) {
		t.Parallel()
		api := &API{
			host: "test_failed_host",
		}
		indexCh := make(chan int)
		resultCh := make(chan *apiResult)
		api.wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go api.getCharacterListJob(ctx, indexCh, resultCh)
		indexCh <- 0
		apiResult := <-resultCh
		require.Error(t, apiResult.err)
		require.Empty(t, apiResult.listIDs)
		require.Equal(t, 0, apiResult.index)
		close(indexCh)
		api.wg.Wait()
	})
	t.Run("test_get_index_success", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleJsonFromMarvel))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		indexCh := make(chan int)
		resultCh := make(chan *apiResult)
		api.wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go api.getCharacterListJob(ctx, indexCh, resultCh)
		indexCh <- 0
		apiResult := <-resultCh
		require.NoError(t, apiResult.err)
		require.Len(t, apiResult.listIDs, 1)
		require.Equal(t, 0, apiResult.index)
		close(indexCh)
		api.wg.Wait()
	})
}

func TestDoGetListCharacters(t *testing.T) {
	t.Parallel()
	t.Run("host_error", func(t *testing.T) {
		t.Parallel()
		api := &API{
			host: "test_failed_host",
		}
		list, total, err := api.DoGetListCharacters(0, API_LIMIT)
		require.Error(t, err)
		require.Empty(t, list)
		require.Zero(t, total)
	})
	t.Run("api_result_json_error", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler("invalid json"))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		list, total, err := api.DoGetListCharacters(0, API_LIMIT)
		require.Error(t, err)
		require.Empty(t, list)
		require.Zero(t, total)
	})
	t.Run("api_result_invalid_code", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(`{"code": 409}`))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		list, total, err := api.DoGetListCharacters(0, API_LIMIT)
		require.Error(t, err)
		require.Empty(t, list)
		require.Zero(t, total)
	})
	t.Run("api_result_invalid_data", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(`{"code": 200}`))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		list, total, err := api.DoGetListCharacters(0, API_LIMIT)
		require.Error(t, err)
		require.Empty(t, list)
		require.Zero(t, total)
	})
	t.Run("api_result_success", func(t *testing.T) {
		t.Parallel()
		testServer, err := test.NewTestServer(test.NewMockHandler(test.SampleJsonFromMarvel))
		require.NoError(t, err)
		api := &API{
			host: test.GetHost(testServer.URL),
		}
		list, total, err := api.DoGetListCharacters(0, API_LIMIT)
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, 1, total)
	})
}
