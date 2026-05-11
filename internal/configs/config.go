package configs

import (
	"flag"
	"os"
)

const (
	DefaultEmptySting    = ""
	DefaultAddr          = "localhost:8080"
	DefaultBaseURL       = "http://localhost:8080"
	DefaultLevelLog      = "info"
	DefaultMigrationPath = "./migrations"
)

type StorageConfig struct {
	Path           string
	PathDB         string
	PathMigrations string
}

type LogConfig struct {
	Level string
}

type Config struct {
	Addr    string
	BaseURL string
	Log     LogConfig
	Storage StorageConfig
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
