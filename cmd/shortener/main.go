package main

import (
	"github.com/eshadow1/shortener/internal/handler"
	"github.com/eshadow1/shortener/internal/repository"
	"github.com/eshadow1/shortener/internal/service"

	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := repository.NewMemoryRepository()
	s := service.NewShortenerService(r)
	h := handler.NewHandler(s)

	rs := chi.NewRouter()
	rs.Get("/{id}", h.GetOrigin)
	rs.Post("/", h.PostCreate)

	err := http.ListenAndServe(`:8080`, rs)
	if err != nil {
		panic(err)
	}
}
