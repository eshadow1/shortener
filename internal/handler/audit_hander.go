package handler

import (
	"context"
	"net/http"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/go-chi/chi/v5/middleware"
)

type ctxKey struct{}

// SetAuditData сохраняет данные аудита в контексте запроса.
func SetAuditData(r *http.Request, eventType model.ActionType, url string) {
	if userID, ok := r.Context().Value(model.UserIDContextKey).(string); ok {
		ctx := context.WithValue(r.Context(), ctxKey{}, model.NewEvent(eventType, &userID, url))
		*r = *r.WithContext(ctx)
	} else {
		ctx := context.WithValue(r.Context(), ctxKey{}, model.NewEvent(eventType, nil, url))
		*r = *r.WithContext(ctx)
	}
}

// GetAuditData извлекает данные аудита из контекста.
func GetAuditData(r *http.Request) *model.Event {
	v, ok := r.Context().Value(ctxKey{}).(*model.Event)
	if !ok {
		loggers.Log.Errorf("audit data not found in request context")
	}
	return v
}

// AuditBroker - брокер, который осуществляет рассылку.
type AuditBroker interface {
	Notify(model.Event)
}

// AuditMiddleware оборачивает хэндлер и после его успешной работы
// отправляет событие аудита во все зарегистрированные приёмники.
func AuditMiddleware(broker AuditBroker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrap := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrap, r)

			if wrap.Status() != http.StatusConflict && wrap.Status() >= http.StatusBadRequest {
				return
			}

			data := GetAuditData(r)
			if data == nil {
				return
			}

			broker.Notify(*data)
		})
	}
}
