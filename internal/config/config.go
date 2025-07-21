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
	HOST                     string `default:"0.0.0.0" required:"true"`
	PORT                     string `default:"4002" required:"true"`
	ENV                      string `default:"development" required:"true"`
	ServiceName              string `default:"subtitler-api" required:"true" splits_words:"true"`
	Debug                    bool   `default:"false"`
	OpenSubtitlesApiKey      string `required:"true" split_words:"true"`
	OpenSubtitlesApiUsername string `required:"true" split_words:"true"`
	OpenSubtitlesApiPassword string `required:"true" split_words:"true"`
	OtelEnabled              bool   `required:"true" split_words:"true"`
	OtelEndpoint             string `split_words:"true"`
	LokiEndpoint             string `split_words:"true"`
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
