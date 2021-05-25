package master

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nanmu42/gzip"

	"github.com/fufuok/xy-message-center/common"
	"github.com/fufuok/xy-message-center/conf"
	"github.com/fufuok/xy-message-center/controller"
	"github.com/fufuok/xy-message-center/middleware"
)

// 接口服务
func startWebServer() {
	var app *gin.Engine

	if conf.Config.SYSConf.Debug {
		app = gin.Default()
	} else {
		// 生产环境不记录请求日志, 由相应接口记录
		gin.SetMode(gin.ReleaseMode)
		app = gin.New()
		app.Use(gzip.DefaultHandler().Gin, middleware.RecoveryWithLog(true))
	}

	app = controller.SetupRouter(app)

	common.Log.WithContext("addr", conf.WebServerAddr).Info("Listening and serving HTTP")
	if err := app.Run(conf.WebServerAddr); err != nil {
		log.Fatalln("Failed to start HTTP Server:", err, "\nbye.")
	}
}
