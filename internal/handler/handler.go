package handler

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/go-chi/chi/v5"
)

type service interface {
	CreateShortUrl(context.Context, string) (string, error)
	GetOriginalURL(context.Context, string) (string, error)
}

type handler struct {
	cfg *configs.Config
	s   service
}

func NewHandler(cfg *configs.Config, svc service) *handler {
	return &handler{
		cfg: cfg,
		s:   svc}
}

func (h *handler) PostCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Bad request", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	short, errCreate := h.s.CreateShortUrl(r.Context(), originalURL)
	if errCreate != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(h.cfg.BaseUrl + "/" + short))
	if err != nil {
		http.Error(w, "Internal Server", http.StatusInternalServerError)
		return
	}
}

func (h *handler) GetOrigin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		http.Error(w, "Bad request", http.StatusMethodNotAllowed)
		return
	}

	short := chi.URLParam(r, "shortURL")
	originalURL, errGet := h.s.GetOriginalURL(r.Context(), strings.TrimPrefix(short, "/"))
	if errGet != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
