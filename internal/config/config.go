package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	ENV_PREFFIX = "SA"
)

type Config struct {
	HOST        string `default:"0.0.0.0" required:"true"`
	PORT        string `default:"4002" required:"true"`
	ServiceName string `default:"subtitler-api" required:"true" split_words:"true"`
	Debug       bool   `default:"false"`
}

func New() *Config {
	godotenv.Load()
	cfg, err := Get()
	if err != nil {
		panic(fmt.Errorf("invalid config value(s) in environment file: %w", err))
	}
	return cfg
}

func Get() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process(ENV_PREFFIX, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
