package configs

import (
	"testing"
	"time"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Init(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name                  string
		baseURL               string
		addr                  string
		logLevel              string
		storagePathDB         string
		storagePath           string
		storagePathMigrations string
		authJWTSecret         []byte
		authTokenIssuer       string
		auditFile             string
		auditURL              string
		serviceBufferSizeChan int
		serviceFlushInterval  time.Duration
		serviceBatchSize      int
	}{
		{
			name:                  "success",
			baseURL:               DefaultBaseURL,
			addr:                  DefaultAddr,
			logLevel:              DefaultLevelLog,
			storagePathDB:         DefaultEmptyString,
			storagePath:           DefaultEmptyString,
			storagePathMigrations: DefaultMigrationPath,
			authJWTSecret:         []byte(DefaultEmptyString),
			authTokenIssuer:       DefaultEmptyString,
			auditFile:             DefaultEmptyString,
			auditURL:              DefaultEmptyString,
			serviceBufferSizeChan: DefaultBufferSizeChan,
			serviceBatchSize:      DefaultBatchSize,
			serviceFlushInterval:  DefaultFlushIntervalSecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.Init()

			assert.Equal(t, test.addr, cfg.Addr)
			assert.Equal(t, test.baseURL, cfg.BaseURL)
			assert.Equal(t, test.logLevel, cfg.Log.Level)
			assert.Equal(t, test.storagePathDB, cfg.Storage.PathDB)
			assert.Equal(t, test.storagePath, cfg.Storage.Path)
			assert.Equal(t, test.storagePathMigrations, cfg.Storage.PathMigrations)
			assert.Equal(t, test.authJWTSecret, cfg.Auth.JWTSecret)
			assert.Equal(t, test.authTokenIssuer, cfg.Auth.TokenIssuer)
			assert.Equal(t, test.auditFile, cfg.Audit.File)
			assert.Equal(t, test.auditURL, cfg.Audit.URL)
			assert.Equal(t, test.serviceBufferSizeChan, cfg.Service.BufferSizeChan)
			assert.Equal(t, test.serviceBatchSize, cfg.Service.BatchSize)
			assert.Equal(t, test.serviceFlushInterval, cfg.Service.FlushInterval)
		})
	}
}
