package configs

import (
	"flag"
	"os"
)

const (
	DefaultAddr    = "localhost:8080"
	DefaultBaseUrl = "http://localhost:8080"
)

type Config struct {
	Addr    string
	BaseUrl string
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
}

func (c *Config) parseWithFlag() {
	flag.StringVar(&c.Addr, "a", DefaultAddr, "host:port")
	flag.StringVar(&c.BaseUrl, "b", DefaultBaseUrl, "base url")

	flag.Parse()
}
