// Package httpapi 是 GoTikTok 的 HTTP 接口层，负责路由注册与请求中间件。
// 包名取 httpapi 而非 http，是为了与标准库 net/http 区分，避免导入名冲突。
package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zhirogo/gotiktok/interface/httpapi/handler"
)

// NewRouter 构建并返回配置了 zap 日志与恢复中间件的 Gin 路由。
func NewRouter(logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(zapLogger(logger), zapRecovery(logger))
	r.GET("/ping", handler.Ping)

	return r
}

// zapLogger 记录每个 HTTP 请求的方法、路径、状态码与耗时。
func zapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.Info("http_request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// zapRecovery 捕获 handler 中的 panic，记录日志并返回 500。
func zapRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic_recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
