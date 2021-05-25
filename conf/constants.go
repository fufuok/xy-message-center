package conf

import (
	"time"
)

const (
	CurrentVersion = "1.2.4.21052509"
	LastChange     = "Milestone Version"
	WebServerAddr  = ":27779"
	APPName        = "XY.MessageCenter"
	ProjectName    = "xymessagecenter"

	DDAPPKeyName    = "MSG_DD_APP_KEY"
	DDAppSecretName = "MSG_DD_APP_SECRET"

	// 项目基础密钥 (环境变量名)
	BaseSecretKeyName = "MSG_BASE_SECRET_KEY"
	// 用于解密基础密钥值的密钥 (编译在程序中)
	BaseSecretSalt = "Fufu@msg.Demo"

	// ES 数据分隔符
	ESBodySep = "=-:-="

	// ES 单次批量写入最大条数或最大字节数
	ESPostBatchNum   = 3000
	ESPOSTBatchBytes = 30 << 20

	// Log 缓存表名, 时间格式化
	LogCacheTable = "LOG:15:04"
	// 繁忙的日志限制 (每秒最多 3 个日志)
	LogLimitDuration = time.Second
	LogLimitNum      = 3
	LogDataName      = "LOG_DATA"

	// 文件变化监控时间间隔(分)
	WatcherInterval = 1

	// 钉钉默认文件类型 (KB)
	DDMediaFileType = ".doc.docx.xls.xlsx.ppt.pptx.zip.pdf.rar"
	DDMediaFileSize = 10000
)
