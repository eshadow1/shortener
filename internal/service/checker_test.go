package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCheckerRepository struct {
	mock.Mock
}

func (m *MockCheckerRepository) PingContext(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestCheckerService_CheckDB(t *testing.T) {
	err := loggers.CreateLogger("error")
	require.NoError(t, err)

	tests := []struct {
		name            string
		haveConnection  bool
		errorPing       error
		expectedConnect error
	}{
		{
			name:            "success_connect",
			haveConnection:  true,
			errorPing:       nil,
			expectedConnect: nil,
		},
		{
			name:            "without_connection",
			haveConnection:  false,
			errorPing:       nil,
			expectedConnect: fmt.Errorf("not used database"),
		},
		{
			name:            "unsuccess_connect",
			haveConnection:  true,
			errorPing:       errors.New("test"),
			expectedConnect: fmt.Errorf("not connected to database"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var s *checkerService
			if test.haveConnection {
				mr := new(MockCheckerRepository)
				mr.On("PingContext", t.Context()).Return(test.errorPing)
				s = NewCheckerService(mr)
			} else {
				s = NewCheckerService(nil)
			}

			errConnect := s.ConnectDB(t.Context())
			assert.Equal(t, test.expectedConnect, errConnect)
		})
	}
}
