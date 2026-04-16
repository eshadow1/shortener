package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/go-chi/chi/v5"
)

type service interface {
	CreateShortURL(context.Context, model.OriginalInfo) (model.ShortenInfo, error)
	GetOriginalURL(context.Context, model.ShortenInfo) (model.OriginalInfo, error)
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

	short, errCreate := h.s.CreateShortURL(r.Context(), model.OriginalInfo{OriginalURL: originalURL})
	if errCreate != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(h.cfg.BaseURL + "/" + short.ShortURL))
	if err != nil {
		http.Error(w, "Internal Server", http.StatusInternalServerError)
		return
	}
}

func (h *handler) PostShorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Bad request", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var req model.OriginalInfo
	errUnmarshal := json.Unmarshal(body, &req)
	if errUnmarshal != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	short, errCreate := h.s.CreateShortURL(r.Context(), req)
	if errCreate != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	short.ShortURL = h.cfg.BaseURL + "/" + short.ShortURL

	w.Header().Set("Content-Type", "application/json")

	bodyResponse, errMarshal := json.Marshal(short)
	if errMarshal != nil {
		http.Error(w, "Internal Server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bodyResponse)
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
	originalURL, errGet := h.s.GetOriginalURL(r.Context(), model.ShortenInfo{ShortURL: strings.TrimPrefix(short, "/")})
	if errGet != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
