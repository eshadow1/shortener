package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	CreateShortURL(context.Context, []model.OriginalInfo) ([]model.ShortenInfo, error)
	GetOriginalURL(context.Context, model.ShortenInfo) (model.OriginalInfo, error)
	GetUserURLs(context.Context) ([]model.UserURL, error)
}

type Checker interface {
	ConnectDB(ctx context.Context) error
}

type handler struct {
	cfg *configs.Config
	s   Service
	c   Checker
}

func NewHandler(cfg *configs.Config, svc Service, check Checker) *handler {
	return &handler{
		cfg: cfg,
		s:   svc,
		c:   check,
	}
}

func (h *handler) PostCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	status := http.StatusCreated
	short, errCreate := h.s.CreateShortURL(r.Context(), []model.OriginalInfo{{OriginalURL: originalURL}})
	if errCreate != nil {
		if _, ok := errors.AsType[*model.CustomPostgresError](errCreate); ok {
			status = http.StatusConflict
		} else {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}
	shortURL, errJoin := url.JoinPath(h.cfg.BaseURL, short[0].ShortURL)
	if errJoin != nil {
		loggers.Log.Errorf("Error joining short url: %v", errJoin)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		loggers.Log.Errorf("Error writing response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handler) PostShorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req model.OriginalInfo
	errUnmarshal := json.Unmarshal(body, &req)
	if errUnmarshal != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	status := http.StatusCreated
	shorts, errCreate := h.s.CreateShortURL(r.Context(), []model.OriginalInfo{req})
	if errCreate != nil {
		if _, ok := errors.AsType[*model.CustomPostgresError](errCreate); ok {
			status = http.StatusConflict
		} else {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	short := shorts[0]
	var errJoin error
	short.ShortURL, errJoin = url.JoinPath(h.cfg.BaseURL, short.ShortURL)
	if errJoin != nil {
		loggers.Log.Errorf("Error joining short url: %v", errJoin)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	bodyResponse, errMarshal := json.Marshal(map[string]string{"result": short.ShortURL})
	if errMarshal != nil {
		loggers.Log.Errorf("Error marshaling response: %v", errMarshal)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	_, err = w.Write(bodyResponse)
	if err != nil {
		loggers.Log.Errorf("Error writing response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handler) PostShortenBatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req []model.OriginalInfo
	errUnmarshal := json.Unmarshal(body, &req)
	if errUnmarshal != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	shorts, errCreate := h.s.CreateShortURL(r.Context(), req)
	if errCreate != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	for i, short := range shorts {
		var errJoin error
		shorts[i].ShortURL, errJoin = url.JoinPath(h.cfg.BaseURL, short.ShortURL)
		if errJoin != nil {
			loggers.Log.Errorf("Error joining short url: %v", errJoin)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	bodyResponse, errMarshal := json.Marshal(shorts)
	if errMarshal != nil {
		loggers.Log.Errorf("Error marshaling response: %v", errMarshal)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bodyResponse)
	if err != nil {
		loggers.Log.Errorf("Error writing response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handler) GetOrigin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusMethodNotAllowed)
		return
	}

	short := chi.URLParam(r, "shortURL")
	originalURL, errGet := h.s.GetOriginalURL(r.Context(), model.ShortenInfo{ShortURL: strings.TrimPrefix(short, "/")})
	if errGet != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	userURLs, errGetURLs := h.s.GetUserURLs(r.Context())
	if errGetURLs != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(userURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for i, userUrl := range userURLs {
		var errJoin error
		userURLs[i].ShortURL, errJoin = url.JoinPath(h.cfg.BaseURL, userUrl.ShortURL)
		if errJoin != nil {
			loggers.Log.Errorf("Error joining short url: %v", errJoin)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	bodyResponse, errMarshal := json.Marshal(userURLs)
	if errMarshal != nil {
		loggers.Log.Errorf("Error marshaling response: %v", errMarshal)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, errBody := w.Write(bodyResponse)
	if errBody != nil {
		loggers.Log.Errorf("Error writing response: %v", errBody)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handler) GetCheckDB(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	errConnect := h.c.ConnectDB(r.Context())
	if errConnect != nil {
		loggers.Log.Errorf("Error connecting to DB: %v", errConnect)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		loggers.Log.Errorf("Error writing response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
