package configs

import (
	"flag"
	"os"
)

const (
	DefaultAddr     = "localhost:8080"
	DefaultBaseUrl  = "http://localhost:8080"
	DefaultLevelLog = "info"
)

type Config struct {
	Addr     string
	BaseUrl  string
	LevelLog string
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
		c.BaseUrl = baseUrl
	}

	if levelLog := os.Getenv("LEVEL_LOG"); levelLog != "" {
		c.LevelLog = levelLog
	}
}

func (c *Config) parseWithFlag() {
	flag.StringVar(&c.Addr, "a", DefaultAddr, "host:port")
	flag.StringVar(&c.BaseUrl, "b", DefaultBaseUrl, "base url")
	flag.StringVar(&c.LevelLog, "l", DefaultLevelLog, "level log")

	flag.Parse()
}
