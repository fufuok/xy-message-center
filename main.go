package main

import (
	"github.com/zh-five/xdaemon"

	"github.com/fufuok/xy-message-center/conf"
	"github.com/fufuok/xy-message-center/master"
)

func main() {
	if !conf.Config.SYSConf.Debug {
		xdaemon.NewDaemon(conf.LogDaemon).Run()
	}

	master.Start()
	master.Watcher()

	select {}
}
