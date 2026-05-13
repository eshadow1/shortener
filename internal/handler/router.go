package handler

import (
	"net/http"
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	timeoutRequest = 5 * time.Second
)

type RouterHandler interface {
	GetOrigin(w http.ResponseWriter, r *http.Request)
	PostCreate(w http.ResponseWriter, r *http.Request)
	PostShorten(w http.ResponseWriter, r *http.Request)
	PostShortenBatch(w http.ResponseWriter, r *http.Request)
	GetCheckDB(w http.ResponseWriter, r *http.Request)
	GetUserURLs(w http.ResponseWriter, r *http.Request)
}

func InitRouter(cfg *configs.Config, h RouterHandler) *chi.Mux {
	rs := chi.NewRouter()
	rs.Use(LoggerMiddleware(), GzipMiddleware(), AuthMiddleware(&cfg.Auth), middleware.Timeout(timeoutRequest))

	rs.Route("/", func(r chi.Router) {
		r.Post("/", h.PostCreate)
		r.Get("/{shortURL}", h.GetOrigin)
		rs.Get("/ping", h.GetCheckDB)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", h.PostShorten)
				r.Post("/batch", h.PostShortenBatch)
			})
			r.Get("/user/urls", h.GetUserURLs)
		})
	})

	return rs
}
