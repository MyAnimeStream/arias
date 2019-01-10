package arias

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"github.com/micro/go-config/source/flag"
)

type Config struct {
	ServerAddr string
}

func LoadConfig(configFile string) (c Config, err error) {
	goConfig := config.NewConfig()
	err = goConfig.Load(
		file.NewSource(file.WithPath(configFile)),
		env.NewSource(),
		flag.NewSource(),
	)

	if err != nil {
		return
	}

	err = goConfig.Scan(&c)
	return
}
