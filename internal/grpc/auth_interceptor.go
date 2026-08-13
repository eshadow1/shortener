package grpcserver

import (
	"context"
	"strings"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/eshadow1/shortener/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor создает unary interceptor для проверки JWT-токена в gRPC.
func AuthInterceptor(cfg *configs.AuthConfig) grpc.UnaryServerInterceptor {
	worker := service.NewJWTWorker(cfg)

	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		var token string

		if ok {
			authHeaders := md.Get("authorization")
			if len(authHeaders) > 0 {
				token = strings.TrimPrefix(authHeaders[0], "Bearer ")
			}
		}

		setResponseToken := func(newToken string) {
			header := metadata.Pairs("authorization", "Bearer "+newToken)
			if err := grpc.SetHeader(ctx, header); err != nil {
				loggers.Log.Errorf("failed to set header: %v", err)
			}
		}

		var uid string

		if token == "" {
			newUID, newToken, errCreate := worker.CreateNewJWT()
			if errCreate != nil {
				return nil, status.Error(codes.Internal, "failed to create user token")
			}

			uid = newUID
			setResponseToken(newToken)
		} else {
			claims, errValidate := worker.ValidateJWT(token, cfg.JWTSecret)
			if errValidate != nil {
				newUID, newToken, errCreate := worker.CreateNewJWT()
				if errCreate != nil {
					return nil, status.Error(codes.Internal, "failed to create user token")
				}

				uid = newUID
				setResponseToken(newToken)
			} else {
				if claims.UserID == "" {
					return nil, status.Error(codes.Unauthenticated, "invalid user ID in token")
				}
				uid = claims.UserID
			}
		}

		ctx = context.WithValue(ctx, model.UserIDContextKey, uid)

		return handler(ctx, req)
	}
}
