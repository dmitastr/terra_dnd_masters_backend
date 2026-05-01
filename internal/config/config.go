package config

import (
	"fmt"
	"os"
)

type ConfigProvider interface {
	GetAddress() string
}

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) GetAddress() string {
	var addr string
	if addr = os.Getenv("PORT"); addr == "" {
		addr = "8080"
	}
	return fmt.Sprintf(":%s", addr)
	// panic("no address were given")
}
