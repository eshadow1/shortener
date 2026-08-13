// Package grpcserver предоставляет описание работы по протоколу gRPC для приложения,
// Пакет инициализирует GRPCServer и описывает его методы.
package grpcserver

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/eshadow1/shortener/gen/pb"
	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/handler"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/eshadow1/shortener/internal/service"
)

type GRPCServer struct {
	pb.UnimplementedShortenerServiceServer
	cfg *configs.Config
	s   handler.Service
}

func NewGRPCServer(cfg *configs.Config, s handler.Service) *GRPCServer {
	return &GRPCServer{
		cfg: cfg,
		s:   s,
	}
}

// ShortenURL обрабатывает запросы с JSON-телом для создания короткого URL.
func (s *GRPCServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	originalInfo := model.OriginalInfo{OriginalURL: req.GetUrl()}
	shorts, errCreate := s.s.CreateShortURL(ctx, []model.OriginalInfo{originalInfo})

	if errCreate != nil {
		if _, ok := errors.AsType[*model.CustomPostgresError](errCreate); ok {
			if len(shorts) > 0 {
				shortURL, _ := url.JoinPath(s.cfg.BaseURL, shorts[0].ShortURL)
				return nil, status.Errorf(codes.AlreadyExists, "URL already exists: %s", shortURL)
			}
			return nil, status.Error(codes.AlreadyExists, "URL already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create short URL")
	}

	if len(shorts) == 0 {
		return nil, status.Error(codes.Internal, "no short URL returned")
	}

	shortURL, errJoin := url.JoinPath(s.cfg.BaseURL, shorts[0].ShortURL)
	if errJoin != nil {
		return nil, status.Error(codes.Internal, "failed to join URL")
	}

	return &pb.URLShortenResponse{Result: shortURL}, nil
}

// ExpandURL обрабатывает запросы для замены короткого URL на оригинальный.
func (s *GRPCServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	shortInfo := model.ShortenInfo{ShortURL: strings.TrimPrefix(req.GetId(), "/")}
	originalInfo, errGet := s.s.GetOriginalURL(ctx, shortInfo)

	if errGet != nil {
		if errors.Is(errGet, service.ErrorDeleteShortURL) {
			return nil, status.Error(codes.NotFound, "URL has been deleted")
		}
		return nil, status.Error(codes.InvalidArgument, "URL not found")
	}

	return &pb.URLExpandResponse{Result: originalInfo.OriginalURL}, nil
}

// ListUserURLs обрабатывает запросы для получения списка всех URL-адресов,
// созданных текущим пользователем, и возвращает их в формате JSON.
func (s *GRPCServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userURLs, errGetURLs := s.s.GetUserURLs(ctx)
	if errGetURLs != nil {
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	resp := &pb.UserURLsResponse{
		Url: make([]*pb.URLData, 0, len(userURLs)),
	}

	for _, u := range userURLs {
		shortURL, errJoin := url.JoinPath(s.cfg.BaseURL, u.ShortURL)
		if errJoin != nil {
			return nil, status.Error(codes.Internal, "failed to join URL")
		}
		resp.Url = append(resp.Url, &pb.URLData{
			ShortUrl:    shortURL,
			OriginalUrl: u.OriginalURL,
		})
	}

	return resp, nil
}

// InitGRPCServer создает и настраивает gRPC сервер.
func InitGRPCServer(ctx context.Context, cfg *configs.Config, s handler.Service) (*grpc.Server, net.Listener, error) {
	lc := &net.ListenConfig{}

	lis, err := lc.Listen(ctx, "tcp", cfg.GRPCAddr)
	if err != nil {
		return nil, nil, err
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(AuthInterceptor(&cfg.Auth)),
	)

	pb.RegisterShortenerServiceServer(grpcSrv, NewGRPCServer(cfg, s))

	return grpcSrv, lis, nil
}
