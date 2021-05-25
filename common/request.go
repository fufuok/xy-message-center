package common

import (
	"bytes"
	"time"

	"github.com/fufuok/utils"
	"github.com/imroc/req"
	"github.com/muesli/cache2go"

	"github.com/fufuok/xy-message-center/conf"
)

var bodySep = []byte(conf.ESBodySep)

// 发送上一分钟数据到 ES
func SendLastMinuteData(key string, apiUrl string) {
	var bodyBuf bytes.Buffer
	i := 0
	cacheName := time.Now().Add(-1 * time.Minute).Format(key)
	cache2go.Cache(cacheName).Foreach(func(_ interface{}, item *cache2go.CacheItem) {
		bodyBuf.Write(utils.GetBytes(item.Data()))
		bodyBuf.Write(bodySep)
		i = i + 1
		// 按内容大小或条数分送发送
		if i%conf.ESPostBatchNum == 0 || bodyBuf.Len() > conf.ESPOSTBatchBytes {
			_, _ = req.Post(apiUrl, req.BodyJSON(&bodyBuf), conf.ReqUserAgent)
			bodyBuf.Reset()
			i = 0
		}
	})
	if i > 0 {
		_, _ = req.Post(apiUrl, req.BodyJSON(&bodyBuf), conf.ReqUserAgent)
	}
}
