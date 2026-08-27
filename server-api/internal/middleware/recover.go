package middleware

import (
	"fmt"
	"megin/pkg/context/api"
	"megin/pkg/logger"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 全局错误拦截器,可以在业务中随便panic,但是不建议这么干
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				if c.IsAborted() {
					c.Status(200)
				}

				result := api.Result[any]{
					Code:   api.STATUS_SERVER_ERROR,
					Trace:  normalizeStackLines(stack),
				}

				switch errType := err.(type) {
				case string:
					result.Message = errType
				case error:
					result.Message = errType.Error()
				default:
					result.Message = "Unkonw Error"
				}

				logger.Error("请求错误", zap.Any("error", err))
				logger.Error(fmt.Sprintf("请求错误堆栈:\n%s", stack))
				c.JSON(http.StatusOK, result)
			}
		}()
		c.Next()
	}
}

// 规范化堆栈行，去掉首尾空白，便于接口直接展示。
func normalizeStackLines(stack string) []string {
	lines := strings.Split(strings.TrimRight(stack, "\n"), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		result = append(result, strings.TrimSpace(line))
	}
	return result
}
