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
	GetCheckDB(w http.ResponseWriter, r *http.Request)
	GetUserURLs(w http.ResponseWriter, r *http.Request)
	PostCreate(w http.ResponseWriter, r *http.Request)
	PostShorten(w http.ResponseWriter, r *http.Request)
	PostShortenBatch(w http.ResponseWriter, r *http.Request)
	DeleteUserURLs(w http.ResponseWriter, r *http.Request)
}

// InitRouter инициализирует и настраивает HTTP-маршрутизатор (chi.Mux) для приложения.
func InitRouter(cfg *configs.Config, h RouterHandler, a AuditBroker) *chi.Mux {
	audit := AuditMiddleware(a)

	rs := chi.NewRouter()
	rs.Use(LoggerMiddleware(), GzipMiddleware(), AuthMiddleware(&cfg.Auth), middleware.Timeout(timeoutRequest))
	rs.Route("/", func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			r.Use(audit)
			r.Post("/", h.PostCreate)
			r.Get("/{shortURL}", h.GetOrigin)
		})
		r.Get("/ping", h.GetCheckDB)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Route("/", func(r chi.Router) {
					r.Use(audit)
					r.Post("/", h.PostShorten)
				})
				r.Post("/batch", h.PostShortenBatch)
			})
			r.Get("/user/urls", h.GetUserURLs)
			r.Delete("/user/urls", h.DeleteUserURLs)
		})
	})

	return rs
}
