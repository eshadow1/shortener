package main

import (
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/handler"
	"github.com/eshadow1/shortener/internal/repository"
	"github.com/eshadow1/shortener/internal/service"

	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	defaultReadTimeout  = 15 * time.Second
	defaultWriteTimeout = 15 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

func main() {
	cfg := configs.NewConfig()
	cfg.ParseWithFlag()

	r := repository.NewMemoryRepository()
	s := service.NewShortenerService(r)
	h := handler.NewHandler(cfg, s)

	rs := chi.NewRouter()
	rs.Get("/{shortURL}", h.GetOrigin)
	rs.Post("/", h.PostCreate)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      rs,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
