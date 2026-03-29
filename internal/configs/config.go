package configs

import (
	"flag"
	"strings"
)

const (
	DefaultAddr    = "localhost:8080"
	DefaultBaseUrl = "http://localhost:8080/qsd54gFg"
)

type Config struct {
	Addr    string
	BaseUrl string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) GetBaseDelta() string {
	delta := strings.Split(c.BaseUrl, "/")
	return delta[len(delta)-1]
}

func (c *Config) ParseWithFlag() {
	flag.StringVar(&c.Addr, "a", DefaultAddr, "host:port")
	flag.StringVar(&c.BaseUrl, "b", DefaultBaseUrl, "base url")

	flag.Parse()
}
