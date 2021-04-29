/**
Test Utils for marvel api
*/
package test

import (
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/gorilla/mux"
)

func NewTestServer(handler http.Handler) (*httptest.Server, error) {
	ts := httptest.NewUnstartedServer(handler)
	ts.Start()
	return ts, nil
}

func NewMockHandler(json string) http.Handler {
	m := &MockHandler{
		json: json,
	}
	router := mux.NewRouter()
	router.Path("/v1/public/characters").HandlerFunc(m.GetListCharacters)
	router.Path("/v1/public/characters/{id:[0-9]+}").HandlerFunc(m.GetCharacterInfo)
	return router
}

type MockHandler struct {
	json string
}

func (h *MockHandler) GetListCharacters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(h.json))
}
func (h *MockHandler) GetCharacterInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(h.json))
}

type mockFlakyError struct {
	*MockHandler
	flakyIndex int
}

func NewMockFlakyHandler(json string, index int) http.Handler {
	m := &mockFlakyError{
		MockHandler: &MockHandler{
			json: json,
		},
		flakyIndex: index,
	}
	router := mux.NewRouter()
	router.Path("/v1/public/characters").HandlerFunc(m.GetListCharacters)
	router.Path("/v1/public/characters/{id:[0-9]+}").HandlerFunc(m.GetCharacterInfo)
	return router
}

func (h *mockFlakyError) GetListCharacters(w http.ResponseWriter, r *http.Request) {
	o := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(o)
	if err == nil && offset == h.flakyIndex {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("invalid json"))
		return
	}
	h.MockHandler.GetListCharacters(w, r)
}
