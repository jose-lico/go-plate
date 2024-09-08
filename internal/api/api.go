package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type APIServer struct {
}

func NewAPIServer() *APIServer {
	return &APIServer{}
}

func (s *APIServer) Run() error {
	router := chi.NewRouter()

	router.Mount("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	}))

	return http.ListenAndServe("localhost:8000", router)
}
