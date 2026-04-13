package handler

import (
	"net/http"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/go-chi/chi/v5"
)

type routerHandler interface {
	GetOrigin(w http.ResponseWriter, r *http.Request)
	PostCreate(w http.ResponseWriter, r *http.Request)
	PostShorten(w http.ResponseWriter, r *http.Request)
}

func InitRouter(h routerHandler) *chi.Mux {
	rs := chi.NewRouter()
	rs.Use(loggers.RequestLogger())
	rs.Get("/{shortURL}", h.GetOrigin)
	rs.Post("/", h.PostCreate)
	rs.Post("/api/shorten", h.PostShorten)
	return rs
}
