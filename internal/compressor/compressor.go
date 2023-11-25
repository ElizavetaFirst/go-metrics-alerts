package compressor

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GzipGinRequestMiddleware(c *gin.Context) {
	if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
		reader, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode gzip request"})
			c.Abort()
			return
		}
		c.Request.Body = reader
	}
	c.Next()
}
