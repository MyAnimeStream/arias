package arias

import (
	"errors"
	"github.com/micro/go-config"
	"github.com/micro/go-config/source"
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
		ServerAddr:  ":80",
		Aria2Addr:   "ws://localhost:6800/jsonrpc",
		StorageType: "s3",
	}
}

func LoadConfig(configFile string) (c Config, err error) {
	c = defaultConfig()

	goConfig := config.NewConfig()

	sources := []source.Source{env.NewSource()}
	if configFile != "" {
		sources = append([]source.Source{file.NewSource(file.WithPath(configFile))}, sources...)
	}

	err = goConfig.Load(sources...)
	if err != nil {
		return
	}

	err = goConfig.Scan(&c)
	return
}

func (c *Config) Check() error {
	if c == nil {
		return errors.New("config is nil")
	}

	return nil
}
