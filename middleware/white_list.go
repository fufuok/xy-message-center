package middleware

import (
	"net"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-message-center/common"
	"github.com/fufuok/xy-message-center/conf"
)

// 接口白名单检查
func CheckWhiteList(asAPI bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(conf.Config.DD.WhiteList) > 0 {
			clientIP := c.ClientIP()
			ip := net.ParseIP(clientIP)
			forbidden := true
			for ipNet := range conf.Config.DD.WhiteList {
				if ipNet.Contains(ip) {
					forbidden = false
					break
				}
			}

			if forbidden {
				msg := "非法来访: " + clientIP
				common.LogLimit.Warn(msg)
				if asAPI {
					APIFailure(c, msg, nil)
				} else {
					TxtMsg(c, msg)
				}

				return
			}
		}

		c.Next()
	}
}
