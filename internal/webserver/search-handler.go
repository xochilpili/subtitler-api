package webserver

import (
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
