package master

import (
	"log"
	"time"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-message-center/common"
	"github.com/fufuok/xy-message-center/conf"
	"github.com/fufuok/xy-message-center/service/dingtalk"
)

// 监听程序二进制变化(重启)和配置文件(热加载)
func Watcher() {
	mainFile := utils.Executable()
	if mainFile == "" {
		log.Fatalln("Failed to initialize Watcher: miss executable", "\nbye.")
	}

	md5Main, _ := utils.MD5Sum(mainFile)
	md5Conf, _ := utils.MD5Sum(conf.ConfigFile)

	common.Log.
		WithContext("main", mainFile, "config", conf.ConfigFile).
		Info("Watching")

	go func() {
		for range time.Tick(time.Duration(conf.Config.SYSConf.WatcherInterval) * time.Minute) {
			// 程序二进制变化时重启
			md5New, _ := utils.MD5Sum(mainFile)
			if md5New != md5Main {
				md5Main = md5New
				common.Log.Warn(">>>>>>> restart main <<<<<<<")
				restartChan <- true
				continue
			}
			// 配置文件变化时热加载
			md5New, _ = utils.MD5Sum(conf.ConfigFile)
			if md5New != md5Conf {
				md5Conf = md5New
				oldDDAccessToken := conf.Config.DD.AccessToken
				if err := conf.LoadConf(); err != nil {
					common.Log.Error("reload config err: ", err)
					continue
				}

				// 重启程序指令
				if conf.Config.SYSConf.RestartMain {
					common.Log.Warn(">>>>>>> restart main(config) <<<<<<<")
					restartChan <- true
					continue
				}

				// 日志配置更新
				_ = common.InitLogger()

				// 重新刷新 DD AccessToken
				if _, err := dingtalk.RefreshAccessToken(); err != nil {
					// 配置有误, 保持原来的 AccessToken
					conf.Config.DD.AccessToken = oldDDAccessToken
					common.Log.Error("refresh dingtalk access_token err: ", err)
				}

				common.Log.Warn("white list: ", conf.Config.DD.WhiteList)
				common.Log.Warn(">>>>>>> reload config <<<<<<<")
				reloadChan <- true
			}
		}
	}()
}
