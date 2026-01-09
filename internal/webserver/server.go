package webserver

import (
	"context"
	"io"
	"net/http"
	"time"

	ginlogger "github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/xochilpili/subtitler-api/internal/config"
	"github.com/xochilpili/subtitler-api/internal/models"
	"github.com/xochilpili/subtitler-api/internal/providers"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
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

	// metrics instruments
	reqCounter   otelmetric.Int64Counter
	durHistogram otelmetric.Float64Histogram
}

func New(config *config.Config, logger *zerolog.Logger) *WebServer {
	ginger := gin.New()
	ginger.Use(gin.Recovery())

	ginger.Use(otelgin.Middleware(
		config.ServiceName,
		otelgin.WithFilter(func(r *http.Request) bool {
			return r.URL.Path != "/ping"
		}),
	))

	ginger.Use(ginlogger.SetLogger(
		ginlogger.WithSkipPath([]string{"/ping"}),
		ginlogger.WithLogger(func(ctx *gin.Context, l zerolog.Logger) zerolog.Logger {
			return logger.Output(gin.DefaultWriter).With().Logger()
		}),
	))

	// Initialize metrics instruments (safe if no MeterProvider configured)
	meter := otel.Meter(config.ServiceName)
	reqCounter, _ := meter.Int64Counter(
		"http.server.requests",
		otelmetric.WithDescription("Number of HTTP server requests"),
	)
	durHistogram, _ := meter.Float64Histogram(
		"http.server.request.duration",
		otelmetric.WithDescription("Duration of HTTP server requests in seconds"),
		otelmetric.WithUnit("s"),
	)

	// Attach metrics middleware
	ginger.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start).Seconds()
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		attrs := []attribute.KeyValue{
			attribute.String("http.route", route),
			attribute.String("http.method", c.Request.Method),
			attribute.Int("http.status_code", c.Writer.Status()),
		}
		ctx := c.Request.Context()
		reqCounter.Add(ctx, 1, otelmetric.WithAttributes(attrs...))
		durHistogram.Record(ctx, dur, otelmetric.WithAttributes(attrs...))
	})

	httpSrv := &http.Server{
		Addr:    config.HOST + ":" + config.PORT,
		Handler: ginger,
	}

	manager := providers.New(config, logger)
	srv := &WebServer{
		config:       config,
		logger:       logger,
		Web:          httpSrv,
		ginger:       ginger,
		manager:      manager,
		reqCounter:   reqCounter,
		durHistogram: durHistogram,
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
		download.GET("/:provider/:subtitleId", w.Download)
	}
}
