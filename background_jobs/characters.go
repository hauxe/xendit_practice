/**
Job running asynchronously to update character list
*/
package jobs

import (
	"encoding/json"
	"log"
	"time"

	"github.com/hauxe/xendit_pratice/cacher"
	"github.com/hauxe/xendit_pratice/marvel"
)

// StartUpdateCharacterListJob this function fork a nnew goroutine for periodically check for new character
// in marvel api and start update the character list
func StartUpdateCharacterListJob(shutdown <-chan struct{},
	tick time.Duration,
	c cacher.Cacher,
	cacheKey string,
	marvelAPI *marvel.API,
) {
	go func() {
		ticker := time.NewTicker(tick)
		defer ticker.Stop()
		for {
			select {
			case <-shutdown:
				// got a shutdown signal, quit the process
				return
			case <-ticker.C:
				shouldUpdate, err := checkMarvelUpdate(c, cacheKey, marvelAPI)
				if err != nil {
					// if error occur we shouldn't interupt the process
					// just log for monitoring/alerting
					log.Println("[AsyncJob]Check Marvel Update got error", err)
					// becareful of break
					// this break the select statement
					// if somehow the code is copied for refactor it may break unexpected statement: for, if
					break
				}
				if shouldUpdate {
					// get the new list and update cache
					list, err := marvelAPI.GetAllCharacters()
					if err != nil {
						// just log for monitoring/alerting
						log.Println("[AsyncJob]Check Marvel Update got error", err)
						// becareful of break
						break
					}
					b, err := json.Marshal(&list)
					if err != nil {
						// just log for monitoring/alerting
						log.Println("[AsyncJob]Check Marvel Update got error", err)
						// becareful of break
						break
					}
					c.Set(cacheKey, string(b))
				}
			}
		}
	}()
}

func checkMarvelUpdate(c cacher.Cacher, cacheKey string, marvelAPI *marvel.API) (bool, error) {
	// get from cache, if it doesn't got cached, request to get the list
	v, found := c.Get(cacheKey)
	if !found {
		return true, nil
	}
	var list []int
	err := json.Unmarshal([]byte(v), &list)
	if err != nil {
		// some how we store a corrupted data?
		// log error
		log.Println("[AsyncJob]Cache a corrupted character list", v)
		return true, nil
	}
	// get first api of marvel to get the total data
	_, total, err := marvelAPI.DoGetListCharacters(0, 1)
	if err != nil {
		return false, err
	}
	if total != len(list) {
		return true, nil
	}
	return false, nil
}
