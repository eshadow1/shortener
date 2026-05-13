package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDContextKey contextKey = "user_id"
)

type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
