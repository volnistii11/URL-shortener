package middlewares

import (
	"io"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (r *gzipWriter) Write(b []byte) (int, error) {
	return r.Writer.Write(b)
}
