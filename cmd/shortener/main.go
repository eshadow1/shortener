package main

import (
	"github.com/eshadow1/shortener/internal/handler"
	"github.com/eshadow1/shortener/internal/repository"
	"github.com/eshadow1/shortener/internal/service"

	"net/http"
)

func main() {
	r := repository.NewMemoryRepository()
	s := service.NewShortenerService(r)
	h := handler.NewHandler(s)

	mux := http.NewServeMux()
	mux.HandleFunc(`/{id}`, h.GetOrigin)
	mux.HandleFunc(`/`, h.PostCreate)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
