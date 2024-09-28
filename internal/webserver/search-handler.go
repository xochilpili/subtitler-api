package webserver

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (w *WebServer) SearchAll(c *gin.Context) {
	query := c.Query("term")
	if query == "" {
		c.JSON(http.StatusBadRequest, &gin.H{"mesasge": "error", "error": "bad request"})
		return
	}

	subtitles := w.manager.Search(c.Request.Context(), query)
	c.JSON(http.StatusOK, &gin.H{"message": "ok", "total": len(subtitles), "data": subtitles})
}

func (w *WebServer) Download(c *gin.Context) {
	subtitleId := c.Param("subtitleId")
	if subtitleId == "" {
		c.JSON(http.StatusBadRequest, &gin.H{"message": "error", "error": "bad request"})
		return
	}
	w.logger.Info().Msgf("downloading subtitle: %s", subtitleId)
	body, filename, contentType, err := w.manager.Download(c.Request.Context(), subtitleId)
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
