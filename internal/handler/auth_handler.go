// Package handler предоставляет HTTP-хендлеры для обработки сервиса сокращения ссылок, логирование запросов,
// аудит успешных вызовов, кодирования и запросов аутентификации.
// Пакет инкапсулирует работу с HTTP-протоколом.
package handler

import (
	"context"
	"net/http"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/eshadow1/shortener/internal/service"
)

type JWTWorker interface {
	GetUID(context.Context, string, []byte) (model.TokenAuth, error)
}

// AuthMiddleware создает middleware для проверки JWT-токена и авторизации пользователя.
func AuthMiddleware(cfg *configs.AuthConfig) func(http.Handler) http.Handler {
	worker := service.NewJWTWorker(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cookie, errCookie := r.Cookie(service.CookieName)
			var token string
			if errCookie != nil || cookie.Value == "" {
				token = ""
			} else {
				token = cookie.Value
			}

			tokenID, errGetUID := worker.GetUID(ctx, token, cfg.JWTSecret)
			if errGetUID != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if tokenID.IsNewToken {
				setAuthCookie(w, tokenID.Token)
			}

			if tokenID.UID == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, model.UserIDContextKey, tokenID.UID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// setAuthCookie устанавливает в HTTP-ответ куку с JWT-токеном.
func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     service.CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   service.CookieMaxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
