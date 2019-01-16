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

	DefaultBucket string
	// AllowBucketOverride specifies whether the requester can override the bucket to upload to
	AllowBucketOverride bool

	// AllowNoName specifies whether the requester may omit the name from the request.
	// Arias would use the name of the downloaded file in that case
	AllowNoName bool
}

func defaultConfig() Config {
	return Config{
		ServerAddr: ":7200",
		Aria2Addr:  "ws://localhost:6800/jsonrpc",
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

	if c.DefaultBucket == "" {
		return errors.New("default bucket must be specified")
	}

	return nil
}
