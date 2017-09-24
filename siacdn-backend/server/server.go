package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

var ErrNotImplemented error = errors.New("Endpoint has not yet been implemented")

type HTTPDServer struct {
	mux http.Handler
}

func NewHTTPDServer() (*HTTPDServer, error) {
	s := &HTTPDServer{}

	router := s.makeRouter()
	s.mux = cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(http.Handler(router))

	return s, nil
}

func (s *HTTPDServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *HTTPDServer) Close() error {
	return nil
}

func (s *HTTPDServer) makeRouter() *httprouter.Router {
	r := httprouter.New()
	r.GET("/", s.handleGetIndex)
	/*
		r.GET("/auth", s.handleGetAuth)
		r.POST("/auth", s.handleCreateAuthToken)
		r.POST("/accounts", s.handleCreateAccount)
		r.GET("/accounts/id/:id", s.handleGetAccount)
	*/
	r.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.JsonErr(w, "Not found")
	})
	return r
}

func (s *HTTPDServer) JsonErr(w http.ResponseWriter, msg string) {
	log.Println(msg)
	s.Json(w, map[string]string{"error": msg})
}

func (s *HTTPDServer) Json(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if encoded, err := json.Marshal(data); err == nil {
		w.Write(encoded)
	} else {
		w.Write([]byte("{\"error\": \"Sorry, there was an error" +
			", and then an error getting the error message from it.\"}"))
	}
}

func (s *HTTPDServer) limitOffset(r *http.Request) (limit int, offset int) {
	// Set defaults
	limit = 20
	offset = 0

	// Now pull out args from the query
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Try to parse limit
	lim, err := strconv.ParseInt(limitStr, 10, 32)
	if err == nil {
		limit = int(lim)
	} else {
		//log.Println("Could not parse limit: " + limitStr + " : " + err.Error())
	}

	// Try to parse offset
	offs, err := strconv.ParseInt(offsetStr, 10, 32)
	if err == nil {
		offset = int(offs)
	} else {
		//log.Println("Could not parse offset: " + offsetStr + " : " + err.Error())
	}

	return
}