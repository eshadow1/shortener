package configs

import (
	"flag"
	"os"
)

const (
	DefaultAddr          = "localhost:8080"
	DefaultBaseURL       = "http://localhost:8080"
	DefaultLevelLog      = "info"
	DefaultStoragePath   = "./storage.txt"
	DefaultStoragePathDB = ""
)

type StorageConfig struct {
	Path   string
	PathDB string
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

	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		c.Addr = addr
	}

	if baseUrl := os.Getenv("BASE_URL"); baseUrl != "" {
		c.BaseURL = baseUrl
	}

	if levelLog := os.Getenv("LEVEL_LOG"); levelLog != "" {
		c.Log.Level = levelLog
	}

	if storagePath := os.Getenv("FILE_STORAGE_PATH"); storagePath != "" {
		c.Storage.Path = storagePath
	}

	if pathDB := os.Getenv("DATABASE_DSN"); pathDB != "" {
		c.Storage.PathDB = pathDB
	}
}

func (c *Config) parseWithFlag() {
	flag.StringVar(&c.Addr, "a", DefaultAddr, "host:port")
	flag.StringVar(&c.BaseURL, "b", DefaultBaseURL, "base url")
	flag.StringVar(&c.Log.Level, "l", DefaultLevelLog, "level log")
	flag.StringVar(&c.Storage.Path, "f", DefaultStoragePath, "file storage path")
	flag.StringVar(&c.Storage.PathDB, "d", DefaultStoragePathDB, "file storage path")

	flag.Parse()
}
