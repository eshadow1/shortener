// Package service предоставляет реализацию бизнес-логики приложения.
package service

import (
	"errors"
	"net/http"
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// CookieName определяет имя HTTP-куки, в которой хранится JWT-токен аутентификации.
	CookieName = "auth_token"
	// CookieMaxAge определяет срок жизни JWT-токена и соответствующей куки в секундах (30 дней).
	CookieMaxAge = 30 * 24 * 60 * 60 // 30 дней
)

var (
	// ErrInvalidToken возвращается при обнаружении недействительного, просроченного или неправильно подписанного JWT-токена.
	ErrInvalidToken = errors.New("invalid token")
)

type jwtWorker struct {
	cfg *configs.AuthConfig
}

// NewJWTWorker создает и возвращает новый экземпляр компонента для работы с JWT-токенами.
func NewJWTWorker(cfg *configs.AuthConfig) *jwtWorker {
	return &jwtWorker{
		cfg: cfg,
	}
}

// CreateNewJWTForUser генерирует новый идентификатор пользователя, создает для него JWT-токен.
func (jw *jwtWorker) CreateNewJWTForUser(w http.ResponseWriter) (string, error) {
	uid := jw.GenerateUserID()
	token, err := jw.CreateJWT(uid, jw.cfg.JWTSecret)
	if err != nil {
		return uid, err
	}
	jw.SetAuthCookie(w, token)
	return uid, nil
}

// CreateJWT формирует и подписывает новый JWT-токен с утверждениями.
func (jw *jwtWorker) CreateJWT(userID string, secret []byte) (string, error) {
	claims := model.UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jw.cfg.TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(CookieMaxAge) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateJWT выполняет разбор и проверку подписи JWT-токена,
// возвращая извлеченные утверждения пользователя в случае успешной валидации.
func (*jwtWorker) ValidateJWT(tokenString string, secret []byte) (*model.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*model.UserClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// SetAuthCookie устанавливает в HTTP-ответ куку с JWT-токеном.
func (*jwtWorker) SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// GenerateUserID генерирует новый уникальный идентификатор пользователя в формате UUID.
func (*jwtWorker) GenerateUserID() string {
	return uuid.NewString()
}
