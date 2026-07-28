package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtWorker_NewJWTWorker(t *testing.T) {
	cfg := &configs.AuthConfig{
		JWTSecret:   []byte("test-secret-key"),
		TokenIssuer: "test-issuer",
	}
	worker := NewJWTWorker(cfg)
	assert.NotNil(t, worker)
	assert.Equal(t, cfg, worker.cfg)
}

func TestJwtWorker_GenerateUserID(t *testing.T) {
	cfg := &configs.AuthConfig{
		JWTSecret:   []byte("test-secret-key"),
		TokenIssuer: "test-issuer",
	}
	worker := NewJWTWorker(cfg)

	userID := worker.GenerateUserID()
	_, err := uuid.Parse(userID)
	require.NoError(t, err)

	userID2 := worker.GenerateUserID()
	assert.NotEqual(t, userID, userID2)
}

func TestJwtWorker_CreateJWT(t *testing.T) {
	cfg := &configs.AuthConfig{
		JWTSecret:   []byte("test-secret-key"),
		TokenIssuer: "test-issuer",
	}
	worker := NewJWTWorker(cfg)

	tests := []struct {
		name    string
		userID  string
		secret  []byte
		wantErr bool
	}{
		{
			name:    "success",
			userID:  "user-123",
			secret:  cfg.JWTSecret,
			wantErr: false,
		},
		{
			name:    "empty userID",
			userID:  "",
			secret:  cfg.JWTSecret,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := worker.CreateJWT(tt.userID, tt.secret)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			claims, err := worker.ValidateJWT(token, cfg.JWTSecret)
			require.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, cfg.TokenIssuer, claims.Issuer)
			assert.NotZero(t, claims.IssuedAt)
			assert.NotZero(t, claims.ExpiresAt)
		})
	}
}

func TestJwtWorker_SetAuthCookie(t *testing.T) {
	cfg := &configs.AuthConfig{
		JWTSecret:   []byte("test-secret-key"),
		TokenIssuer: "test-issuer",
	}
	worker := NewJWTWorker(cfg)

	rec := httptest.NewRecorder()
	token := "test-jwt-token"
	worker.SetAuthCookie(rec, token)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, CookieName, cookie.Name)
	assert.Equal(t, token, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, CookieMaxAge, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.False(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestJwtWorker_CreateNewJWTForUser(t *testing.T) {
	cfg := &configs.AuthConfig{
		JWTSecret:   []byte("test-secret-key"),
		TokenIssuer: "test-issuer",
	}
	worker := NewJWTWorker(cfg)

	rec := httptest.NewRecorder()

	userID, err := worker.CreateNewJWTForUser(rec)
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	_, err = uuid.Parse(userID)
	require.NoError(t, err)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, CookieName, cookie.Name)
	assert.NotEmpty(t, cookie.Value)

	claims, err := worker.ValidateJWT(cookie.Value, cfg.JWTSecret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, cfg.TokenIssuer, claims.Issuer)
}
