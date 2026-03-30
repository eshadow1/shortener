package configs

import (
	"flag"
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

func (c *Config) ParseWithFlag() {
	flag.StringVar(&c.Addr, "a", DefaultAddr, "host:port")
	flag.StringVar(&c.BaseUrl, "b", DefaultBaseUrl, "base url")

	flag.Parse()
}
