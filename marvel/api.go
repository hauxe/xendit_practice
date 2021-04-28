package marvel

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	API_LIMIT = 100
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
func NewAPI(publicKey, privateKey string) *API {
	return &API{
		host:          "gateway.marvel.com",
		apiPublicKey:  publicKey,
		apiPrivateKey: privateKey,
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
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
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
	return apiResult.Data.Results[0], nil
}

// GetListCharacters get all characters
func (api *API) GetAllCharacters() (result []int, _ error) {
	// get first index to determine the total count
	list, total, err := api.doGetListCharacters(0)
	if err != nil {
		return nil, fmt.Errorf("get api index 0 error: %w", err)
	}
	result = append(result, list...)
	if len(list) >= total {
		return
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
	num := total / 100
	if total%100 != 0 {
		num += 1
	}
	i := 1
	for {
		select {
		case indexCh <- i:
			i++
			if i >= num {
				indexCh = nil
			}
		case apiResult := <-resultCh:
			if apiResult.err != nil {
				return nil, fmt.Errorf("get api index %d error: %w", apiResult.index, apiResult.err)
			}
			result = append(result, apiResult.listIDs...)
			if len(result) >= total {
				return
			}
		}
	}
}

func (api *API) getCharacterListJob(ctx context.Context, indexCh <-chan int, resultCh chan<- *apiResult) {
	defer api.wg.Done()
	result := new(apiResult)
	for i := range indexCh {
		result.index = i
		list, _, err := api.doGetListCharacters(i)
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

func (api *API) doGetListCharacters(index int) ([]int, int, error) {
	offset := index * API_LIMIT
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
	query.Set("limit", strconv.Itoa(API_LIMIT))
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
