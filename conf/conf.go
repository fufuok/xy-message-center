package conf

import (
	"path/filepath"

	"github.com/fufuok/utils"
	"github.com/imroc/req"
)

// 运行绝对路径
var RootPath = utils.ExecutableDir(true)

// 配置文件绝对路径
var FilePath = filepath.Join(RootPath, "..", "etc")

// 默认配置文件路径
var ConfigFile = filepath.Join(FilePath, ProjectName+".json")

// 日志路径
var LogDir = filepath.Join(RootPath, "..", "log")

// 守护日志
var LogDaemon = filepath.Join(LogDir, "daemon.log")

// 所有配置
var Config tJSONConf

// 请求名称
var ReqUserAgent = req.Header{"User-Agent": APPName + "/" + CurrentVersion}
