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

type JWTWorker interface {
	GetUID(context.Context, string, []byte) (model.TokenAuth, error)
}

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

		tokenID, errGetUID := worker.GetUID(ctx, token, cfg.JWTSecret)
		if errGetUID != nil {
			return nil, status.Error(codes.Internal, "failed to create user token")
		}

		if tokenID.IsNewToken {
			updateHeaders(ctx, tokenID.Token)
		}

		if tokenID.UID == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid user ID in token")
		}

		ctx = context.WithValue(ctx, model.UserIDContextKey, tokenID.UID)
		return handler(ctx, req)
	}
}

// updateHeaders устанавливает в GRPC-ответ с JWT-токеном.
func updateHeaders(ctx context.Context, token string) {
	header := metadata.Pairs("authorization", "Bearer "+token)
	if err := grpc.SetHeader(ctx, header); err != nil {
		loggers.Log.Errorf("failed to set header: %v", err)
	}
}
