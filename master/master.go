package master

import (
	"context"
	"os"
	"time"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-message-center/common"
	"github.com/fufuok/xy-message-center/conf"
	"github.com/fufuok/xy-message-center/service"
)

var (
	// 重启信号
	restartChan = make(chan bool)
	// 配置重载信息
	reloadChan = make(chan bool)
)

func Start() {
	go func() {
		go startWebServer()
		go startLogAPI()

		for {
			// 获取远程配置
			ctx, cancel := context.WithCancel(context.Background())
			go startRemoteConf(ctx)

			select {
			case <-restartChan:
				// 强制退出, 由 Daemon 重启程序
				common.Log.Warn("restart <-restartChan")
				os.Exit(0)
			case <-reloadChan:
				cancel()
				common.Log.Warn("reload <-reloadChan")
			}
		}
	}()
}

// 每分钟写一次日志
func startLogAPI() {
	for range time.Tick(time.Minute) {
		if conf.Config.SYSConf.LogAPI != "" {
			// 心跳日志
			common.LogCache(utils.MustJSON(map[string]interface{}{
				"type":          "heartbeat",
				"internal_ipv4": service.InternalIPv4,
				"external_ipv4": service.ExternalIPv4,
				"time":          time.Now().Format(time.RFC3339),
			}))
			common.SendLastMinuteData(conf.LogCacheTable, conf.Config.SYSConf.LogAPI)
		}
	}
}
