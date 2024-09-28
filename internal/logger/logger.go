package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/xochilpili/subtitler-api/internal/config"
	"github.com/xochilpili/subtitler-api/internal/version"
)

func New(config *config.Config) *zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("service", config.ServiceName).Str("version", version.VERSION).Logger()
	zerolog.DefaultContextLogger = &logger
	return &logger
}
