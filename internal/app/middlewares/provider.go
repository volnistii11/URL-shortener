package middlewares

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"strings"
	"time"
)

type MiddlewareProvider interface {
	LogHTTPHandler() gin.HandlerFunc
	GZIPHandler() gin.HandlerFunc
}

func NewMiddlewareProvider(logger *zap.Logger) MiddlewareProvider {
	return &middleware{
		logger: logger,
	}
}

type middleware struct {
	logger *zap.Logger
}

func (m *middleware) LogHTTPHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := &loggingResponseWriter{
			ResponseWriter: ctx.Writer,
			responseData:   responseData,
		}
		ctx.Writer = lw

		uri := ctx.Request.RequestURI
		method := ctx.Request.Method
		ctx.Next()
		duration := time.Since(start)
		m.logger.Sugar().Infow("request data",
			"uri", uri,
			"method", method,
			"duration", duration)
		m.logger.Sugar().Infow("response data",
			"status", responseData.status,
			"size", responseData.size,
		)
	}
}

func (m *middleware) GZIPHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
			ctx.Next()
			return
		}

		if !strings.Contains(ctx.GetHeader("Content-Type"), "application/json") && !strings.Contains(ctx.GetHeader("Content-Type"), "text/html") {
			ctx.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(ctx.Writer, gzip.BestSpeed)
		if err != nil {
			io.WriteString(ctx.Writer, err.Error())
			return
		}
		defer gz.Close()

		ctx.Header("Content-Encoding", "gzip")
		ctx.Writer = &gzipWriter{
			ResponseWriter: ctx.Writer,
			Writer:         gz,
		}
		ctx.Next()
	}
}
