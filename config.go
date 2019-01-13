package arias

import (
	"errors"
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
)

type Config struct {
	ServerAddr string
	Aria2Addr  string

	StorageType string
}

func defaultConfig() Config {
	return Config{
		ServerAddr:  "localhost:8080",
		Aria2Addr:   "ws://localhost:6800/jsonrpc",
		StorageType: "google",
	}
}

func LoadConfig(configFile string) (c Config, err error) {
	c = defaultConfig()

	goConfig := config.NewConfig()
	err = goConfig.Load(
		file.NewSource(file.WithPath(configFile)),
		env.NewSource(),
	)
	if err != nil {
		return
	}

	err = goConfig.Scan(&c)
	return
}

func (c *Config) Check() error {
	if c == nil {
		return errors.New("Config is nil")
	}

	return nil
}
