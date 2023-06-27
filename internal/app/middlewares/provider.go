package middlewares

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MiddlewareProvider interface {
	LogHTTPHandler() gin.HandlerFunc
	GZIPHandler() gin.HandlerFunc
	Auth() gin.HandlerFunc
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
		if !strings.Contains(ctx.GetHeader("Content-Encoding"), "gzip") {
			ctx.Next()
			return
		}
		gz, err := gzip.NewReader(ctx.Request.Body)
		if err != nil {
			io.WriteString(ctx.Writer, err.Error())
			return
		}
		fmt.Println("gz", gz)
		ctx.Request.Body = gz
		ctx.Request.Body.Close()
		defer gz.Close()

		if !strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
			ctx.Next()
			return
		}
		gzResponse, errResponse := gzip.NewWriterLevel(ctx.Writer, gzip.BestSpeed)
		if errResponse != nil {
			io.WriteString(ctx.Writer, errResponse.Error())
			return
		}
		defer gzResponse.Close()

		ctx.Header("Content-Encoding", "gzip")
		ctx.Writer = &gzipWriter{
			ResponseWriter: ctx.Writer,
			Writer:         gzResponse,
		}

		ctx.Next()
	}
}

func (m *middleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Cookie("cookie_auth")
		if err != nil {
			tokenString, _ := BuildJWTString(rand.Intn(1000))
			ctx.SetCookie("cookie_auth", tokenString, 3600, "/", "localhost", false, false)
			ctx.Next()
			return
		}

		userID := GetUserID(cookie)
		if userID == -1 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		customData := &customData{UserId: userID}
		ctx.Writer = &authingResponseWriter{
			ResponseWriter: ctx.Writer,
			customData:     customData,
		}
		ctx.Set("user_id", userID)
		ctx.Next()
	}
}
