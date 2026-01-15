package config

import (
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"strings"
)

const DefaultSecretKey = "your-secret-key-change-this-in-production"

type Config struct {
	Address string `env:"RUN_ADDRESS"`

	DatabaseDNS    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

	SecretKey string `env:"SECRET_KEY"`
}

func InitConfig() *Config {
	flags := Flags{}
	flags.Init()

	cfg := Config{
		Address:        flags.address,
		DatabaseDNS:    flags.dbDNS,
		AccrualAddress: flags.accrualAddress,
		SecretKey:      DefaultSecretKey,
	}
	cfg.parseEnv()

	return &cfg
}

func (cfg *Config) parseEnv() {
	err := env.Parse(cfg)
	if err != nil {
		logger.Log.Warn("Getting an error while parsing the configuration", zap.String("customerror", err.Error()))
	}
}

func (cfg *Config) GetAccrualAddressWithProtocol() string {
	http := "http://"
	https := "https://"

	if strings.Contains(cfg.AccrualAddress, https) || strings.Contains(cfg.AccrualAddress, http) {
		return cfg.AccrualAddress
	}
	return http + cfg.AccrualAddress
}
