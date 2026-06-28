package configs

import (
	"flag"
	"os"
	"strconv"
	"time"
)

const (
	DefaultEmptySting          = ""
	DefaultAddr                = "localhost:8080"
	DefaultBaseURL             = "http://localhost:8080"
	DefaultLevelLog            = "info"
	DefaultMigrationPath       = "./migrations"
	DefaultBufferSizeChan      = 100
	DefaultBatchSize           = 10
	DefaultFlushIntervalSecond = 15 * time.Second
)

type StorageConfig struct {
	Path           string
	PathDB         string
	PathMigrations string
}

type LogConfig struct {
	Level string
}

type AuthConfig struct {
	JWTSecret   []byte
	TokenIssuer string
}

type ServiceConfig struct {
	BufferSizeChan int
	BatchSize      int
	FlushInterval  time.Duration
}

type AuditConfig struct {
	File string
	URL  string
}

type Config struct {
	Addr    string
	BaseURL string
	Log     LogConfig
	Storage StorageConfig
	Auth    AuthConfig
	Service ServiceConfig
	Audit   AuditConfig
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Init() {
	c.parseWithFlag()

	c.Addr = c.updateEnv("SERVER_ADDRESS", c.Addr)
	c.BaseURL = c.updateEnv("BASE_URL", c.BaseURL)

	c.Log.Level = c.updateEnv("LOG_LEVEL", c.Log.Level)

	c.Storage.Path = c.updateEnv("FILE_STORAGE_PATH", c.Storage.Path)
	c.Storage.PathDB = c.updateEnv("DATABASE_DSN", c.Storage.PathDB)
	c.Storage.PathMigrations = c.updateEnv("MIGRATION_PATH", c.Storage.PathMigrations)

	c.Auth.JWTSecret = []byte(c.updateEnv("JWT_SECRET", string(c.Auth.JWTSecret)))
	c.Auth.TokenIssuer = c.updateEnv("TOKEN_ISSUER", c.Auth.TokenIssuer)

	c.Audit.File = c.updateEnv("AUDIT_FILE", c.Audit.File)
	c.Audit.URL = c.updateEnv("AUDIT_URL", c.Audit.URL)

	if bufferSizeChan, ok := os.LookupEnv("BUFFER_SIZE_CHAN"); ok {
		var errConv error
		c.Service.BufferSizeChan, errConv = strconv.Atoi(bufferSizeChan)
		if errConv != nil {
			c.Service.BufferSizeChan = DefaultBufferSizeChan
		}
	} else {
		c.Service.BufferSizeChan = DefaultBufferSizeChan
	}

	if batchSize, ok := os.LookupEnv("BATCH_SIZE"); ok {
		var errConv error
		c.Service.BatchSize, errConv = strconv.Atoi(batchSize)
		if errConv != nil {
			c.Service.BatchSize = DefaultBatchSize
		}
	} else {
		c.Service.BatchSize = DefaultBatchSize
	}

	if flushInterval, ok := os.LookupEnv("FLUSH_INTERVAL"); ok {
		temp, errConv := strconv.Atoi(flushInterval)
		if errConv != nil {
			c.Service.FlushInterval = DefaultFlushIntervalSecond
		} else {
			c.Service.FlushInterval = time.Duration(temp) * time.Second
		}
	} else {
		c.Service.FlushInterval = DefaultFlushIntervalSecond
	}
}

func (c *Config) parseWithFlag() {
	flag.StringVar(&c.Addr, "a", DefaultAddr, "host:port")
	flag.StringVar(&c.BaseURL, "b", DefaultBaseURL, "base url")
	flag.StringVar(&c.Log.Level, "l", DefaultLevelLog, "level log")
	flag.StringVar(&c.Storage.Path, "f", DefaultEmptySting, "file storage path")
	flag.StringVar(&c.Storage.PathDB, "d", DefaultEmptySting, "file storage path")
	flag.StringVar(&c.Storage.PathMigrations, "m", DefaultMigrationPath, "migrations path")
	flag.StringVar(&c.Audit.URL, "audit-url", DefaultEmptySting, "path to audit log file")
	flag.StringVar(&c.Audit.File, "audit-file", DefaultEmptySting, "remote audit server URL")

	flag.Parse()
}

func (*Config) updateEnv(name, defaultValue string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return defaultValue
}
