package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

type MiddlewareProvider interface {
	LogHTTPHandler() gin.HandlerFunc
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

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		gin.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
