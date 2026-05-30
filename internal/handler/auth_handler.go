package handler

import (
	"context"
	"net/http"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/eshadow1/shortener/internal/service"
)

func AuthMiddleware(cfg *configs.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			worker := service.NewJWTWorker(cfg)
			cookie, errCookie := r.Cookie(service.CookieName)
			if errCookie != nil || cookie.Value == "" {
				uid, errCreate := worker.CreateNewJWTForUser(w)
				if errCreate != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				ctx = context.WithValue(ctx, model.UserIDContextKey, uid)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			claims, errValidate := worker.ValidateJWT(cookie.Value, cfg.JWTSecret)
			if errValidate != nil {
				uid, errCreate := worker.CreateNewJWTForUser(w)
				if errCreate != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				ctx = context.WithValue(ctx, model.UserIDContextKey, uid)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if claims.UserID == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, model.UserIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
