package middlewares

import (
	"github.com/gin-gonic/gin"
	"io"
)

type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (r *gzipWriter) Write(b []byte) (int, error) {
	return r.Writer.Write(b)
}
