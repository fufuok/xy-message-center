package common

import (
	"fmt"
	"log"
	"time"

	"github.com/fufuok/gxlog"
	"github.com/fufuok/gxlog/formatter/json"
	"github.com/fufuok/gxlog/formatter/text"
	"github.com/fufuok/gxlog/iface"
	"github.com/fufuok/gxlog/logger"
	"github.com/fufuok/gxlog/writer"
	"github.com/fufuok/gxlog/writer/file"
	"github.com/fufuok/utils/xid"
	"github.com/imroc/req"
	"github.com/muesli/cache2go"

	"github.com/fufuok/xy-message-center/conf"
)

var (
	Log      = gxlog.Logger()
	LogLimit *logger.Logger
)

func init() {
	if err := InitLogger(); err != nil {
		log.Fatalln("Failed to initialize logger:", err, "\nbye.")
	}
}

func InitLogger() error {
	if err := LogConfig(); err != nil {
		return err
	}

	// 带限制的日志记录器
	LogLimit = Log.WithTimeLimit(conf.Config.SYSConf.LogLimitDuration, conf.Config.SYSConf.LogLimitNum)

	return nil
}

// 日志配置
func LogConfig() error {
	Log.SetSlotLevel(logger.Slot0, iface.Level(conf.Config.SYSConf.LogLevel))
	Log.SetSlotFormatter(logger.Slot0, text.New(text.NewConfig()))

	if conf.Config.SYSConf.Debug {
		req.Debug = true
		return nil
	}

	req.Debug = false

	// 启用错误文件日志
	wt, err := file.Open(file.Config{
		Path: conf.LogDir,
		Base: conf.ProjectName,
	})
	if err != nil {
		return fmt.Errorf("logfile path err: %s\nbye.", err)
	}
	Log.CopySlot(logger.Slot1, logger.Slot0)
	Log.SetSlotWriter(logger.Slot1, wt)
	Log.SetSlotLevel(logger.Slot1, iface.Level(conf.Config.SYSConf.LogFileLevel))

	// 发送日志到 API
	Log.SetSlotFormatter(logger.Slot0, json.New(json.NewConfig()))
	Log.SetSlotWriter(logger.Slot0, writer.Func(func(bs []byte, _ *iface.Record) {
		// 配置热重载后可开关此发送
		if conf.Config.SYSConf.LogAPI != "" {
			LogCache(bs)
		}
	}))

	return nil
}

// 日志暂存
func LogCache(bs []byte) {
	key := xid.NewString()
	table := time.Now().Format(conf.LogCacheTable)
	cache := cache2go.Cache(table)
	cache.Add(key, 180*time.Second, bs)
}
