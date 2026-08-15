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

type ShortenerHandler interface {
	GetOrigin(w http.ResponseWriter, r *http.Request)
	GetUserURLs(w http.ResponseWriter, r *http.Request)
	PostShorten(w http.ResponseWriter, r *http.Request)
}

type RouterHandler interface {
	GetCheckDB(w http.ResponseWriter, r *http.Request)
	GetInternalStats(w http.ResponseWriter, r *http.Request)
	PostCreate(w http.ResponseWriter, r *http.Request)
	PostShortenBatch(w http.ResponseWriter, r *http.Request)
	DeleteUserURLs(w http.ResponseWriter, r *http.Request)
}

// InitRouter инициализирует и настраивает HTTP-маршрутизатор (chi.Mux) для приложения.
func InitRouter(cfg *configs.Config, rH RouterHandler, sH ShortenerHandler, a AuditBroker) *chi.Mux {
	audit := AuditMiddleware(a)
	trusted := TrustedSubnetMiddleware(cfg)

	rs := chi.NewRouter()
	rs.Use(LoggerMiddleware(), GzipMiddleware(), AuthMiddleware(&cfg.Auth), middleware.Timeout(timeoutRequest))
	rs.Route("/", func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			r.Use(audit)
			r.Post("/", rH.PostCreate)
			r.Get("/{shortURL}", sH.GetOrigin)
		})
		r.Get("/ping", rH.GetCheckDB)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Route("/", func(r chi.Router) {
					r.Use(audit)
					r.Post("/", sH.PostShorten)
				})
				r.Post("/batch", rH.PostShortenBatch)
			})
			r.Get("/user/urls", sH.GetUserURLs)
			r.Delete("/user/urls", rH.DeleteUserURLs)
			r.Route("/internal/stats", func(r chi.Router) {
				r.Use(trusted)
				r.Get("/", rH.GetInternalStats)
			})
		})
	})

	return rs
}
