package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-message-center/common"
	"github.com/fufuok/xy-message-center/conf"
)

// Web 日志写入 ES
func WebAPILogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		c.Next()

		// 错误日志
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				common.LogLimit.Error(e)
			}
		}

		// 配置热重载后可开关此发送, log 每分钟自动推送
		if conf.Config.SYSConf.LogAPI != "" {
			// 记录日志
			logData := map[string]interface{}{
				"req_time":         start.Format(time.RFC3339),
				"req_method":       c.Request.Method,
				"req_uri":          c.Request.RequestURI,
				"req_proto":        c.Request.Proto,
				"req_ua":           c.Request.UserAgent(),
				"req_referer":      c.Request.Referer(),
				"req_client_ip":    c.ClientIP(),
				"req_body":         c.GetString(conf.LogDataName),
				"resp_status_code": c.Writer.Status(),
				"resp_body_size":   c.Writer.Size(),
				"api_cost_time":    time.Since(start).String(),
			}
			// 写入日志队列
			common.LogCache(utils.MustJSON(logData))
		}
	}
}

// GinRecovery 及日志
// Ref: https://github.com/gin-contrib/zap
func RecoveryWithLog(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					common.LogLimit.
						WithContext("error", err, "request", string(httpRequest)).
						Error("Recovery: ", c.Request.URL.Path)
					// If the connection is dead, we can't write a status to it.
					_ = c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					common.LogLimit.
						WithContext("error", err, "request", string(httpRequest), "stack", string(debug.Stack())).
						Error("Recovery: ", c.Request.URL.Path)
				} else {
					common.LogLimit.
						WithContext("error", err, "request", string(httpRequest)).
						Error("Recovery: ", c.Request.URL.Path)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
