package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-message-center/common"
)

// 通用异常处理
func APIException(c *gin.Context, code int, msg string, data interface{}) {
	if msg == "" {
		msg = "错误的请求"
	}
	c.JSON(code, common.APIFailureData(msg, data))
	c.Abort()
}

// 返回失败, 状态码: 200
func APIFailure(c *gin.Context, msg string, data interface{}) {
	APIException(c, http.StatusOK, msg, data)
}

// 返回成功, 状态码: 200
func APISuccess(c *gin.Context, data interface{}, count int) {
	c.JSON(http.StatusOK, common.APISuccessData(data, count))
	c.Abort()
}

// 返回文本消息
func TxtMsg(c *gin.Context, msg string) {
	c.String(http.StatusOK, msg)
	c.Abort()
}
