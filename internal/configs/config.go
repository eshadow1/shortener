package configs

import (
	"flag"
	"os"
	"strconv"
)

const (
	DefaultEmptySting     = ""
	DefaultAddr           = "localhost:8080"
	DefaultBaseURL        = "http://localhost:8080"
	DefaultLevelLog       = "info"
	DefaultMigrationPath  = "./migrations"
	DefaultBufferSizeChan = 1
	DefaultBatchSize      = 1
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
}

type Config struct {
	Addr    string
	BaseURL string
	Log     LogConfig
	Storage StorageConfig
	Auth    AuthConfig
	Service ServiceConfig
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Init() {
	c.parseWithFlag()

	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		c.Addr = addr
	}

	if baseUrl, ok := os.LookupEnv("BASE_URL"); ok {
		c.BaseURL = baseUrl
	}

	if levelLog, ok := os.LookupEnv("LEVEL_LOG"); ok {
		c.Log.Level = levelLog
	}

	if storagePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		c.Storage.Path = storagePath
	}

	if pathDB, ok := os.LookupEnv("DATABASE_DSN"); ok {
		c.Storage.PathDB = pathDB
	}

	if pathMigration, ok := os.LookupEnv("MIGRATION_PATH"); ok {
		c.Storage.PathMigrations = pathMigration
	}

	if jwtSecret, ok := os.LookupEnv("JWT_SECRET"); ok {
		c.Auth.JWTSecret = []byte(jwtSecret)
	}

	if tokenIssuer, ok := os.LookupEnv("TOKEN_ISSUER"); ok {
		c.Auth.TokenIssuer = tokenIssuer
	}

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
}

func (c *Config) parseWithFlag() {
	flag.StringVar(&c.Addr, "a", DefaultAddr, "host:port")
	flag.StringVar(&c.BaseURL, "b", DefaultBaseURL, "base url")
	flag.StringVar(&c.Log.Level, "l", DefaultLevelLog, "level log")
	flag.StringVar(&c.Storage.Path, "f", DefaultEmptySting, "file storage path")
	flag.StringVar(&c.Storage.PathDB, "d", DefaultEmptySting, "file storage path")
	flag.StringVar(&c.Storage.PathMigrations, "m", DefaultMigrationPath, "migrations path")

	flag.Parse()
}
