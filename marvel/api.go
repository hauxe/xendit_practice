package marvel

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	API_LIMIT = 100
	API_HOST  = "gateway.marvel.com"
)

// API defines api properties
type API struct {
	host            string
	apiPublicKey    string
	apiPrivateKey   string
	concurrentLimit int
	wg              sync.WaitGroup
}

// NewAPI creates new api object
func NewAPI(host, publicKey, privateKey string) *API {
	if host == "" {
		host = API_HOST
	}
	return &API{
		host:            host,
		apiPublicKey:    publicKey,
		apiPrivateKey:   privateKey,
		concurrentLimit: runtime.NumCPU(),
	}
}

type apiResult struct {
	err     error
	listIDs []int
	index   int
}

type marvelAPIResult struct {
	Code int            `json:"code,omitempty"`
	Data *marvelAPIData `json:"data,omitempty"`
}

type marvelAPIData struct {
	Limit   int                `json:"limit,omitempty"`
	Total   int                `json:"total,omitempty"`
	Count   int                `json:"count,omitempty"`
	Results []*MarvelCharacter `json:"results,omitempty"`
}

type MarvelCharacter struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetCharacterInfo get character info by id
func (api *API) GetCharacterInfo(id int) (*MarvelCharacter, error) {
	ts := time.Now().Unix()
	hash := md5.Sum([]byte(fmt.Sprintf("%d%s%s", ts, api.apiPrivateKey, api.apiPublicKey)))
	u := url.URL{
		Scheme: "http",
		Host:   api.host,
		Path:   "v1/public/characters/" + strconv.Itoa(id),
	}
	query := u.Query()
	query.Set("ts", strconv.FormatInt(ts, 10))
	query.Set("apikey", api.apiPublicKey)
	query.Set("hash", fmt.Sprintf("%x", hash))
	u.RawQuery = query.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("get from %s error: %w", u.String(), err)
	}
	defer resp.Body.Close()
	apiResult := new(marvelAPIResult)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(apiResult); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	if apiResult.Code != 200 {
		return nil, fmt.Errorf("marvel api error, code: %d", apiResult.Code)
	}
	if apiResult.Data == nil {
		return nil, fmt.Errorf("marvel api error, invalid data")
	}
	if len(apiResult.Data.Results) == 0 {
		return nil, fmt.Errorf("no character found for id %d", id)
	}
	info := apiResult.Data.Results[0]
	if info.ID != id {
		return nil, fmt.Errorf("invalid id repsponded: want %d got %d", id, info.ID)
	}
	return info, nil
}

// GetListCharacters get all characters
func (api *API) GetAllCharacters() (result []int, _ error) {
	// get first index to determine the total count
	list, total, err := api.DoGetListCharacters(0, API_LIMIT)
	if err != nil {
		return nil, fmt.Errorf("get api index 0 error: %w", err)
	}
	result = append(result, list...)
	if len(list) >= total {
		return
	}
	if api.concurrentLimit <= 0 {
		api.concurrentLimit = runtime.NumCPU()
	}
	indexCh := make(chan int, api.concurrentLimit)
	resultCh := make(chan *apiResult, api.concurrentLimit)
	// create maximum worker to grab the data concurrently
	api.wg.Add(api.concurrentLimit)
	defer api.wg.Wait()
	defer close(indexCh)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < api.concurrentLimit; i++ {
		go api.getCharacterListJob(ctx, indexCh, resultCh)
	}
	num := total / API_LIMIT
	if total%API_LIMIT != 0 {
		num += 1
	}
	i := 1
	received := 0
	feeder := indexCh
	for {
		select {
		case feeder <- i:
			i++
			if i >= num {
				feeder = nil
			}
		case apiResult := <-resultCh:
			received++
			if apiResult.err != nil {
				return nil, fmt.Errorf("get api index %d error: %w", apiResult.index, apiResult.err)
			}
			result = append(result, apiResult.listIDs...)
			if len(result) >= total {
				return
			}
			if received >= num-1 {
				feeder = indexCh
			}
		}
	}
}

func (api *API) getCharacterListJob(ctx context.Context, indexCh <-chan int, resultCh chan<- *apiResult) {
	defer api.wg.Done()
	result := new(apiResult)
	for i := range indexCh {
		result.index = i
		list, _, err := api.DoGetListCharacters(i, API_LIMIT)
		if err != nil {
			result.err = err
			select {
			case <-ctx.Done():
				return
			case resultCh <- result:
			}
			return
		}
		result.listIDs = list
		select {
		case <-ctx.Done():
			return
		case resultCh <- result:
		}
	}
}

func (api *API) DoGetListCharacters(index int, limit int) ([]int, int, error) {
	offset := index * limit
	ts := time.Now().Unix()
	hash := md5.Sum([]byte(fmt.Sprintf("%d%s%s", ts, api.apiPrivateKey, api.apiPublicKey)))
	u := url.URL{
		Scheme: "http",
		Host:   api.host,
		Path:   "v1/public/characters",
	}
	query := u.Query()
	query.Set("ts", strconv.FormatInt(ts, 10))
	query.Set("apikey", api.apiPublicKey)
	query.Set("hash", fmt.Sprintf("%x", hash))
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	u.RawQuery = query.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, 0, fmt.Errorf("get from %s error: %w", u.String(), err)
	}
	defer resp.Body.Close()
	apiResult := new(marvelAPIResult)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(apiResult); err != nil {
		return nil, 0, fmt.Errorf("invalid response: %w", err)
	}
	if apiResult.Code != 200 {
		return nil, 0, fmt.Errorf("marvel api error, code: %d", apiResult.Code)
	}
	if apiResult.Data == nil {
		return nil, 0, fmt.Errorf("marvel api error, invalid data")
	}
	results := make([]int, 0, apiResult.Data.Count)
	for _, character := range apiResult.Data.Results {
		results = append(results, character.ID)
	}
	return results, apiResult.Data.Total, nil
}
