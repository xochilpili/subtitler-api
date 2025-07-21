package webserver

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xochilpili/subtitler-api/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (w *WebServer) SearchByProvider(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer(w.config.ServiceName)
	ctx, span := tracer.Start(ctx, "Search by provider")
	defer span.End()

	provider := c.Param("provider")
	if provider == "" {
		span.RecordError(errors.New("missing provider"))
		span.SetStatus(http.StatusBadRequest, "missing provider")
		c.JSON(http.StatusBadRequest, &gin.H{"message": "error", "error": "bad request"})
		return
	}
	query := c.Query("term")
	if query == "" {
		span.RecordError(errors.New("missing query term"))
		span.SetStatus(http.StatusBadRequest, "missing query term")
		c.JSON(http.StatusBadRequest, &gin.H{"message": "error", "error": "bad request"})
		return
	}

	ctxSearch, searchSpan := tracer.Start(ctx, "Searching")
	subtitles := w.manager.Search(ctxSearch, provider, query, getPostFilters(c))
	searchSpan.End()

	span.SetAttributes(
		attribute.String("provider", provider),
		attribute.String("query", query),
		attribute.Int("total_result", len(subtitles)),
	)

	c.JSON(http.StatusOK, &gin.H{"message": "ok", "total": len(subtitles), "data": subtitles})
}

func (w *WebServer) SearchAll(c *gin.Context) {
	query := c.Query("term")
	if query == "" {
		c.JSON(http.StatusBadRequest, &gin.H{"mesasge": "error", "error": "bad request"})
		return
	}
	subtitles := w.manager.Search(c.Request.Context(), "", query, getPostFilters(c))
	c.JSON(http.StatusOK, &gin.H{"message": "ok", "total": len(subtitles), "data": subtitles})
}

func (w *WebServer) Download(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, &gin.H{"message": "error", "error": "bad request"})
		return
	}
	subtitleId := c.Param("subtitleId")
	if subtitleId == "" {
		c.JSON(http.StatusBadRequest, &gin.H{"message": "error", "error": "bad request"})
		return
	}
	w.logger.Info().Msgf("downloading subtitle: %s", subtitleId)
	body, filename, contentType, err := w.manager.Download(c.Request.Context(), provider, subtitleId)
	if err != nil {
		c.JSON(http.StatusBadGateway, &gin.H{"message": "error", "error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", contentType)
	_, err = io.Copy(c.Writer, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &gin.H{"message": "error", "error": err.Error()})
		return
	}
	//c.JSON(http.StatusOK, &gin.H{"message": "ok"})
}

func getPostFilters(c *gin.Context) *models.PostFilters {
	postFilter := &models.PostFilters{}
	year := c.Query("year")
	if year != "" {
		y, _ := strconv.Atoi(year)
		postFilter.Year = y
	}
	group := c.Query("group")
	if group != "" {
		postFilter.Group = group
	}
	quality := c.Query("quality")
	if quality != "" {
		postFilter.Quality = quality
	}
	res := c.Query("resolution")
	if res != "" {
		postFilter.Resolution = res
	}
	return postFilter
}
