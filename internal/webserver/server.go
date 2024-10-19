package webserver

import (
	"context"
	"io"
	"net/http"

	ginlogger "github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/xochilpili/subtitler-api/internal/config"
	"github.com/xochilpili/subtitler-api/internal/models"
	"github.com/xochilpili/subtitler-api/internal/providers"
)

type Manager interface {
	Search(ctx context.Context, provider string, query string, filters *models.PostFilters) []models.Subtitle
	Download(ctx context.Context, provider string, subtitleId string) (io.ReadCloser, string, string, error)
}

type WebServer struct {
	config  *config.Config
	logger  *zerolog.Logger
	Web     *http.Server
	ginger  *gin.Engine
	manager Manager
}

func New(config *config.Config, logger *zerolog.Logger) *WebServer {
	ginger := gin.New()
	ginger.Use(gin.Recovery())
	ginger.Use(ginlogger.SetLogger(
		ginlogger.WithSkipPath([]string{"/ping"}),
		ginlogger.WithLogger(func(ctx *gin.Context, l zerolog.Logger) zerolog.Logger {
			return logger.Output(gin.DefaultWriter).With().Logger()
		}),
	))
	httpSrv := &http.Server{
		Addr:    config.HOST + ":" + config.PORT,
		Handler: ginger,
	}

	manager := providers.New(config, logger)
	srv := &WebServer{
		config:  config,
		logger:  logger,
		Web:     httpSrv,
		ginger:  ginger,
		manager: manager,
	}
	srv.loadRoutes()
	return srv
}

func (w *WebServer) loadRoutes() {
	api := w.ginger.Group("/")
	api.GET("/ping", w.PingHandler)
	search := w.ginger.Group("/search")
	{
		// TODO: Add WhisperPath
		search.GET("/all/", w.SearchAll)
		search.GET("/:provider/", w.SearchByProvider)
	}
	download := w.ginger.Group("/download")
	{
		download.GET(":provider/:subtitleId", w.Download)
	}
}
