package conf

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/json"
)

// 接口配置
type tJSONConf struct {
	SYSConf tSYSConf `json:"sys_conf"`
	DD      tDDConf  `json:"dd_conf"`
}

type tSYSConf struct {
	Debug            bool       `json:"debug"`
	LogLevel         int        `json:"log_level"`
	LogFileLevel     int        `json:"log_file_level"`
	LogAPI           string     `json:"log_api"`
	LogLimitTime     int        `json:"log_limit_time"`
	LogLimitNum      int        `json:"log_limit_num"`
	MainConfig       TFilesConf `json:"main_config"`
	RestartMain      bool       `json:"restart_main"`
	WatcherInterval  int        `json:"watcher_interval"`
	BaseSecretValue  string
	LogLimitDuration time.Duration
}

type tDDConf struct {
	AgentID       int                     `json:"agent_id"`
	AppKeyName    string                  `json:"app_key_name"`
	AppSecretName string                  `json:"app_secret_name"`
	WhiteListConf []string                `json:"white_list"`
	MediaConf     map[string]tDDMediaConf `json:"media"`
	WhiteList     map[*net.IPNet]struct{} `json:"-"`
	AppKey        string
	AppSecret     string
	AccessToken   string
}

type tDDMediaConf struct {
	Ext  string `json:"ext"`
	Size int64  `json:"size"`
}

type TFilesConf struct {
	Path            string `json:"path"`
	Method          string `json:"method"`
	SecretName      string `json:"secret_name"`
	API             string `json:"api"`
	Interval        int    `json:"interval"`
	SecretValue     string
	GetConfDuration time.Duration
	ConfigMD5       string
	ConfigVer       time.Time
}

func init() {
	confFile := flag.String("c", ConfigFile, "配置文件绝对路径")
	flag.Parse()
	ConfigFile = *confFile
	if err := LoadConf(); err != nil {
		log.Fatalln("Failed to initialize config:", err, "\nbye.")
	}
}

// 加载配置
func LoadConf() error {
	config, err := readConf()
	if err != nil {
		return err
	}

	Config = *config

	return nil
}

// 读取配置
func readConf() (*tJSONConf, error) {
	body, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	var config *tJSONConf
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, err
	}

	// 基础密钥 Key
	config.SYSConf.BaseSecretValue = utils.GetenvDecrypt(BaseSecretKeyName, BaseSecretSalt)
	if config.SYSConf.BaseSecretValue == "" {
		return nil, fmt.Errorf("%s cannot be empty", BaseSecretKeyName)
	}

	// 日志级别: 1Trace 2Debug 3Info 4Warn 5Error(默认) 6Fatal 7Off
	if config.SYSConf.LogLevel > 7 || config.SYSConf.LogLevel < 1 {
		config.SYSConf.LogLevel = 5
	}

	// 本地文件日志级别
	if config.SYSConf.LogFileLevel > 7 || config.SYSConf.LogFileLevel < 1 {
		config.SYSConf.LogFileLevel = config.SYSConf.LogLevel
	}

	// 繁忙日志限制 (x 秒 n 条)
	if config.SYSConf.LogLimitNum <= 0 || config.SYSConf.LogLimitTime <= 0 {
		config.SYSConf.LogLimitDuration = LogLimitDuration
		config.SYSConf.LogLimitNum = LogLimitNum
	} else {
		config.SYSConf.LogLimitDuration = time.Duration(config.SYSConf.LogLimitTime) * time.Second
	}

	// 钉钉环境变量, 配置中的环境变量名优先
	envName := config.DD.AppKeyName
	if envName == "" {
		envName = DDAPPKeyName
	}
	config.DD.AppKey = utils.GetenvDecrypt(envName, config.SYSConf.BaseSecretValue)
	if config.DD.AppKey == "" {
		return nil, fmt.Errorf("%s cannot be empty", envName)
	}
	envName = config.DD.AppSecretName
	if envName == "" {
		envName = DDAppSecretName
	}
	config.DD.AppSecret = utils.GetenvDecrypt(envName, config.SYSConf.BaseSecretValue)
	if config.DD.AppSecret == "" {
		return nil, fmt.Errorf("%s cannot be empty", envName)
	}

	// IP 白名单
	whiteList := make(map[*net.IPNet]struct{})
	for _, ip := range config.DD.WhiteListConf {
		// 排除空白行, __开头的注释行
		ip = strings.TrimSpace(ip)
		if ip == "" || strings.HasPrefix(ip, "__") {
			continue
		}
		// 补全掩码
		if !strings.Contains(ip, "/") {
			if strings.Contains(ip, ":") {
				ip = ip + "/128"
			} else {
				ip = ip + "/32"
			}
		}
		// 转为网段
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			return nil, err
		}
		whiteList[ipNet] = struct{}{}
	}
	config.DD.WhiteList = whiteList

	// 每次获取远程主配置的时间间隔, < 30 秒则禁用该功能
	if config.SYSConf.MainConfig.Interval > 29 {
		// 远程获取主配置 API, 解密 SecretName
		if config.SYSConf.MainConfig.SecretName != "" {
			config.SYSConf.MainConfig.SecretValue = utils.GetenvDecrypt(config.SYSConf.MainConfig.SecretName,
				config.SYSConf.BaseSecretValue)
			if config.SYSConf.MainConfig.SecretValue == "" {
				return nil, fmt.Errorf("%s cannot be empty", config.SYSConf.MainConfig.SecretName)
			}
		}
		config.SYSConf.MainConfig.GetConfDuration = time.Duration(config.SYSConf.MainConfig.Interval) * time.Second
		config.SYSConf.MainConfig.Path = ConfigFile
	}

	// 文件变化监控时间间隔
	if config.SYSConf.WatcherInterval < 1 {
		config.SYSConf.WatcherInterval = WatcherInterval
	}

	// 钉钉默认媒体文件类型和大小限制
	if _, ok := config.DD.MediaConf["file"]; !ok {
		if config.DD.MediaConf == nil {
			config.DD.MediaConf = make(map[string]tDDMediaConf)
		}
		config.DD.MediaConf["file"] = tDDMediaConf{
			Ext:  DDMediaFileType,
			Size: DDMediaFileSize,
		}
	}

	// 钉钉文件类型前后加 . 便于检查, 大小限制单位转换
	for k, v := range config.DD.MediaConf {
		config.DD.MediaConf[k] = tDDMediaConf{
			Ext:  "." + v.Ext + ".",
			Size: v.Size * 1000,
		}
	}

	return config, nil
}
