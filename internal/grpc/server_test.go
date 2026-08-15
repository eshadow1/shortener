package grpcserver

import (
	"context"
	"errors"
	"testing"

	"github.com/eshadow1/shortener/gen/pb"
	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/handler"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/eshadow1/shortener/internal/service"
	mockhandler "github.com/eshadow1/shortener/mocks/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGRPCServer_ShortenURL(t *testing.T) {
	cfg := &configs.Config{BaseURL: "http://localhost:8080"}

	tests := []struct {
		name        string
		req         *pb.URLShortenRequest
		mockSetup   func(m *mockhandler.MockService)
		wantErr     bool
		wantErrCode codes.Code
		wantResult  string
	}{
		{
			name: "Success",
			req:  &pb.URLShortenRequest{Url: "http://example.com"},
			mockSetup: func(m *mockhandler.MockService) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).
					Return([]model.ShortenInfo{{ShortURL: "/abc"}}, nil)
			},
			wantErr:    false,
			wantResult: "http://localhost:8080/abc",
		},
		{
			name:        "Empty URL",
			req:         &pb.URLShortenRequest{Url: ""},
			mockSetup:   func(*mockhandler.MockService) {},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "CustomPostgresError with shorts > 0",
			req:  &pb.URLShortenRequest{Url: "http://example.com"},
			mockSetup: func(m *mockhandler.MockService) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).
					Return([]model.ShortenInfo{{ShortURL: "/abc"}}, &model.CustomPostgresError{})
			},
			wantErr:     true,
			wantErrCode: codes.AlreadyExists,
		},
		{
			name: "CustomPostgresError with shorts == 0",
			req:  &pb.URLShortenRequest{Url: "http://example.com"},
			mockSetup: func(m *mockhandler.MockService) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).
					Return([]model.ShortenInfo{}, &model.CustomPostgresError{})
			},
			wantErr:     true,
			wantErrCode: codes.AlreadyExists,
		},
		{
			name: "Generic error",
			req:  &pb.URLShortenRequest{Url: "http://example.com"},
			mockSetup: func(m *mockhandler.MockService) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
		{
			name: "Success but empty shorts slice",
			req:  &pb.URLShortenRequest{Url: "http://example.com"},
			mockSetup: func(m *mockhandler.MockService) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).
					Return([]model.ShortenInfo{}, nil)
			},
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mockhandler.MockService)
			tt.mockSetup(mockSvc)

			server := NewGRPCServer(cfg, mockSvc)
			resp, err := server.ShortenURL(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.wantResult, resp.Result)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGRPCServer_ExpandURL(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.URLExpandRequest
		mockSetup   func(m *mockhandler.MockService)
		wantErr     bool
		wantErrCode codes.Code
		wantResult  string
	}{
		{
			name:        "Empty_ID",
			req:         &pb.URLExpandRequest{Id: ""},
			mockSetup:   func(*mockhandler.MockService) {},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "URL_has_been_deleted",
			req:  &pb.URLExpandRequest{Id: "/abc"},
			mockSetup: func(m *mockhandler.MockService) {
				m.On("GetOriginalURL", mock.Anything, model.ShortenInfo{ShortURL: "abc"}).
					Return(model.OriginalInfo{}, service.ErrorDeleteShortURL)
			},
			wantErr:     true,
			wantErrCode: codes.NotFound,
		},
		{
			name: "Generic_error_URL_not_found",
			req:  &pb.URLExpandRequest{Id: "/abc"},
			mockSetup: func(m *mockhandler.MockService) {
				m.On("GetOriginalURL", mock.Anything, model.ShortenInfo{ShortURL: "abc"}).
					Return(model.OriginalInfo{}, errors.New("not found"))
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Success",
			req:  &pb.URLExpandRequest{Id: "/abc"},
			mockSetup: func(m *mockhandler.MockService) {
				m.On("GetOriginalURL", mock.Anything, model.ShortenInfo{ShortURL: "abc"}).
					Return(model.OriginalInfo{OriginalURL: "http://example.com"}, nil)
			},
			wantErr:    false,
			wantResult: "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mockhandler.MockService)
			tt.mockSetup(mockSvc)

			server := NewGRPCServer(&configs.Config{}, mockSvc)
			resp, err := server.ExpandURL(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.wantResult, resp.Result)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGRPCServer_ListUserURLs(t *testing.T) {
	cfg := &configs.Config{BaseURL: "http://localhost:8080"}
	tests := []struct {
		name        string
		mockSetup   func(m *mockhandler.MockService)
		wantErr     bool
		wantErrCode codes.Code
		wantCount   int
		wantResult  []*pb.URLData
	}{
		{
			name: "Service_error",
			mockSetup: func(m *mockhandler.MockService) {
				m.On("GetUserURLs", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
		{
			name: "Success_empty_list",
			mockSetup: func(m *mockhandler.MockService) {
				m.On("GetUserURLs", mock.Anything).
					Return([]model.UserURL{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "Success_non-empty_list",
			mockSetup: func(m *mockhandler.MockService) {
				m.On("GetUserURLs", mock.Anything).
					Return([]model.UserURL{
						{ShortURL: "/abc", OriginalURL: "http://example.com"},
						{ShortURL: "/def", OriginalURL: "http://test.com"},
					}, nil)
			},
			wantErr:   false,
			wantCount: 2,
			wantResult: []*pb.URLData{
				{ShortUrl: "http://localhost:8080/abc", OriginalUrl: "http://example.com"},
				{ShortUrl: "http://localhost:8080/def", OriginalUrl: "http://test.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mockhandler.MockService)
			tt.mockSetup(mockSvc)

			server := NewGRPCServer(cfg, mockSvc)
			resp, err := server.ListUserURLs(context.Background(), &emptypb.Empty{})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Len(t, resp.Url, tt.wantCount)
				assert.ElementsMatch(t, tt.wantResult, resp.Url)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestInitGRPCServer(t *testing.T) {
	type dummyService struct {
		handler.Service
	}

	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "Success_with_auto-assigned_port",
			addr:    "127.0.0.1:0",
			wantErr: false,
		},
		{
			name:    "Failure_with_invalid_address_format",
			addr:    "invalid_address_format",
			wantErr: true,
		},
		{
			name:    "Failure_with_invalid_port_number",
			addr:    "127.0.0.1:999999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &configs.Config{
				GRPCAddr: tt.addr,
				Auth: configs.AuthConfig{
					JWTSecret: []byte("test-secret-for-init"),
				},
			}

			svc := &dummyService{}

			srv, lis, err := InitGRPCServer(t.Context(), cfg, svc)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, srv)
				assert.Nil(t, lis)
			} else {
				require.NoError(t, err)
				require.NotNil(t, srv)
				require.NotNil(t, lis)

				errClose := lis.Close()
				require.NoError(t, errClose)
				srv.Stop()
			}
		})
	}
}
