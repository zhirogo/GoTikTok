// Package handler 提供 HTTP 接口的处理函数。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping 健康检查接口，返回 pong 确认服务存活。
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
