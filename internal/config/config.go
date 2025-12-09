package config

import (
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type Config struct {
	Address string `env:"RUN_ADDRESS"`

	DatabaseDNS string `env:"DATABASE_URI"`
}

func InitConfig() *Config {
	flags := Flags{}
	flags.Init()

	cfg := Config{
		Address:     flags.address,
		DatabaseDNS: flags.dbDNS,
	}
	cfg.parseEnv()

	return &cfg
}

func (cfg *Config) parseEnv() {
	err := env.Parse(cfg)
	if err != nil {
		logger.Log.Warn("Getting an error while parsing the configuration", zap.String("err", err.Error()))
	}
}
