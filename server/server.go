package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	jobs "github.com/hauxe/xendit_pratice/background_jobs"
	"github.com/hauxe/xendit_pratice/cacher"
	"github.com/hauxe/xendit_pratice/marvel"
	"golang.org/x/sync/singleflight"
)

const (
	Characters_Cache_Key     = "characters"
	Character_Info_Cache_Key = "character_info"

	API_PUBLIC_KEY  = "API_PUBLIC_KEY"
	API_PRIVATE_KEY = "API_PRIVATE_KEY"

	UpdateCharacterJobMinute = 24 * 60 // 1 day
)

type Server struct {
	router       *mux.Router
	requestGroup singleflight.Group
	marvelAPI    *marvel.API
	cacher       cacher.Cacher
	shutdown     chan struct{}
}

// NewServer create server
// this will load configuration and init server object
func NewServer() (*Server, error) {
	apiPublicKey, found := os.LookupEnv(API_PUBLIC_KEY)
	if !found {
		return nil, fmt.Errorf("couldn't find environment variable for key %s", API_PUBLIC_KEY)
	}
	apiPrivateKey, found := os.LookupEnv(API_PRIVATE_KEY)
	if !found {
		return nil, fmt.Errorf("couldn't find environment variable for key %s", API_PRIVATE_KEY)
	}
	return &Server{
		router:    mux.NewRouter(),
		marvelAPI: marvel.NewAPI("", apiPublicKey, apiPrivateKey),
		cacher:    cacher.NewCacher(),
		shutdown:  make(chan struct{}),
	}, nil
}

// Start server and listenning on 8080 port
// this function is a blocking function
func (s *Server) Start() {
	defer close(s.shutdown)
	// start async job update character info
	// when the server shutting down, it will cause all async job shutdown too
	jobs.StartUpdateCharacterListJob(s.shutdown,
		time.Minute*time.Duration(UpdateCharacterJobMinute),
		s.cacher,
		Characters_Cache_Key,
		s.marvelAPI,
	)

	// build routes
	s.router.Path("/characters").HandlerFunc(s.GetListCharacters)
	s.router.Path("/characters/{id:[0-9]+}").HandlerFunc(s.GetCharacterInfo)
	fmt.Println("Start Listenning at :8080")
	if err := http.ListenAndServe(":8080", s.router); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) GetListCharacters(w http.ResponseWriter, r *http.Request) {
	v, err, _ := s.requestGroup.Do(Characters_Cache_Key, func() (interface{}, error) {
		// get from cache first
		var list []int
		value, ok := s.cacher.Get(Characters_Cache_Key)
		if ok {
			err := json.Unmarshal([]byte(value), &list)
			if err == nil {
				return list, nil
			}
		}
		list, err := s.marvelAPI.GetAllCharacters()
		if err != nil {
			return nil, err
		}
		v, err := json.Marshal(list)
		if err == nil {
			s.cacher.Set(Characters_Cache_Key, string(v))
		}
		return list, nil
	})
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	list := v.([]int)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&list)
	// log error
	if err != nil {
		log.Println("encode error", err)
	}
}

func (s *Server) GetCharacterInfo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	charID, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("invalid character id"))
		return
	}
	cacheKey := buildCharacterInfoCacheKey(charID)
	v, err, _ := s.requestGroup.Do(cacheKey, func() (interface{}, error) {
		// get from cache first
		info := new(marvel.MarvelCharacter)
		value, ok := s.cacher.Get(cacheKey)
		if ok {
			err := json.Unmarshal([]byte(value), info)
			if err == nil {
				return info, nil
			}
			// log error
			log.Println("decode error", err)
		}
		info, err := s.marvelAPI.GetCharacterInfo(charID)
		if err != nil {
			return nil, err
		}
		v, err := json.Marshal(info)
		if err == nil {
			s.cacher.Set(cacheKey, string(v))
		}
		return info, nil
	})
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	info := v.(*marvel.MarvelCharacter)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(info)
	// log error
	if err != nil {
		log.Println("encode error", err)
	}
}

func buildCharacterInfoCacheKey(id int) string {
	return Character_Info_Cache_Key + "_" + strconv.Itoa(id)
}
