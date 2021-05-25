package master

import (
	"context"
	"reflect"
	"time"

	"github.com/fufuok/xy-message-center/common"
	"github.com/fufuok/xy-message-center/conf"
)

// 初始化获取远端配置
func startRemoteConf(ctx context.Context) {
	// 定时获取远程主配置
	if conf.Config.SYSConf.MainConfig.GetConfDuration > 0 {
		go getRemoteConf(ctx, &conf.Config.SYSConf.MainConfig)
	}
}

// 执行获取远端配置
func getRemoteConf(ctx context.Context, c *conf.TFilesConf) {
	v := reflect.ValueOf(c)
	m := v.MethodByName(c.Method)
	if m.Kind() != reflect.Func {
		common.Log.Error("skip init get remote conf, err func: ", c.Method)
		return
	}

	common.Log.Warnf("start get remote conf: %v, %v, %s for %s, %s", m, &m, c.Method, c.Path, c.GetConfDuration)

	ticker := time.NewTicker(c.GetConfDuration)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			common.Log.Warnf("exit get remote conf: %s for %s", c.Method, c.Path)
			return
		default:
			res := m.Call(nil)
			if !res[0].IsNil() {
				common.Log.Errorf("get remote conf err: %s, method: %s, path: %s", res[0], c.Method, c.Path)
			}
		}
	}
}
