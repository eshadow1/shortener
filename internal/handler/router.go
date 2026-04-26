package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type routerHandler interface {
	GetOrigin(w http.ResponseWriter, r *http.Request)
	PostCreate(w http.ResponseWriter, r *http.Request)
	PostShorten(w http.ResponseWriter, r *http.Request)
	PostShortenBatch(w http.ResponseWriter, r *http.Request)
	GetCheckDB(w http.ResponseWriter, r *http.Request)
}

func InitRouter(h routerHandler) *chi.Mux {
	rs := chi.NewRouter()
	rs.Use(LoggerMiddleware(), GzipMiddleware())

	rs.Route("/", func(r chi.Router) {
		r.Post("/", h.PostCreate)
		r.Get("/{shortURL}", h.GetOrigin)
		rs.Get("/ping", h.GetCheckDB)
		r.Route("/api/shorten", func(r chi.Router) {
			r.Post("/", h.PostShorten)
			r.Post("/batch", h.PostShortenBatch)
		})
	})

	return rs
}
