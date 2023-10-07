package middleware

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/gin-gonic/gin"
)

func RecoverPanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 将异常信息打印到日志
				log.Logger.Error("接口panic错误,err:", err)
				// 返回一个错误响应给客户端
				c.JSON(500, gin.H{
					"error": "服务内部错误",
				})
			}
		}()
		// 继续接受和处理新的请求
		c.Next()
	}
}
