package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-message-center/middleware"
)

func SetupRouter(app *gin.Engine) *gin.Engine {
	v1DD := app.Group("/v1/dd", middleware.CheckWhiteList(true), middleware.WebAPILogger())
	{
		v1DD.POST("/chat/create", V1DDChatCreateHandler)
		v1DD.POST("/chat/send", V1DDChatSendHandler)
		v1DD.POST("/topapi/message", V1DDTopAPIMessageHandler)
		v1DD.POST("/media/upload", V1DDMediaUploadHandler)
	}

	// 健康检查
	app.GET("/heartbeat", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	app.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "PONG")
	})

	// 服务器状态
	app.GET("/sys/status", runningStatusHandler)
	app.GET("/sys/check", middleware.CheckWhiteList(false), func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	// 异常请求
	app.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404")
	})

	app.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "405")
	})

	return app
}
